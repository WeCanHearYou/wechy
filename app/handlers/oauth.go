package handlers

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/enum"

	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/bus"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/jwt"
	"github.com/getfider/fider/app/pkg/log"
	"github.com/getfider/fider/app/pkg/web"
	webutil "github.com/getfider/fider/app/pkg/web/util"
)

// OAuthEcho exchanges OAuth Code for a user profile and return directly to the UI, without storing it
func OAuthEcho() web.HandlerFunc {
	return func(c *web.Context) error {
		provider := c.Param("provider")

		code := c.QueryParam("code")
		if code == "" {
			return c.Redirect("/")
		}

		identifier := c.QueryParam("identifier")
		if identifier == "" || identifier != c.SessionID() {
			log.Warn(c, "OAuth identifier doesn't match with user session ID. Aborting sign in process.")
			return c.Redirect("/")
		}

		rawProfile := &query.GetOAuthRawProfile{Provider: provider, Code: code}
		err := bus.Dispatch(c, rawProfile)
		if err != nil {
			return c.Page(web.Props{
				Title:     "OAuth Test Page",
				ChunkName: "OAuthEcho.page",
				Data: web.Map{
					"err": errors.Cause(err).Error(),
				},
			})
		}

		parseRawProfile := &cmd.ParseOAuthRawProfile{Provider: provider, Body: rawProfile.Result}
		_ = bus.Dispatch(c, parseRawProfile)

		return c.Page(web.Props{
			Title:     "OAuth Test Page",
			ChunkName: "OAuthEcho.page",
			Data: web.Map{
				"body":    rawProfile.Result,
				"profile": parseRawProfile.Result,
			},
		})
	}
}

// OAuthToken exchanges OAuth Code for a user profile
// The user profile is then used to either get an existing user on Fider or creating a new one
// Once Fider user is retrieved/created, an authentication cookie is store in user's browser
func OAuthToken() web.HandlerFunc {
	return func(c *web.Context) error {
		provider := c.Param("provider")
		redirectURL, _ := url.ParseRequestURI(c.QueryParam("redirect"))
		redirectURL.ResolveReference(c.Request.URL)

		code := c.QueryParam("code")
		if code == "" {
			return c.Redirect(redirectURL.String())
		}

		identifier := c.QueryParam("identifier")
		if identifier == "" || identifier != c.SessionID() {
			log.Warn(c, "OAuth identifier doesn't match with user session ID. Aborting sign in process.")
			return c.Redirect(redirectURL.String())
		}

		oauthUser := &query.GetOAuthProfile{Provider: provider, Code: code}
		if err := bus.Dispatch(c, oauthUser); err != nil {
			return c.Failure(err)
		}

		var user *models.User

		userByProvider := &query.GetUserByProvider{Provider: provider, UID: oauthUser.Result.ID}
		err := bus.Dispatch(c, userByProvider)
		user = userByProvider.Result

		if errors.Cause(err) == app.ErrNotFound && oauthUser.Result.Email != "" {
			userByEmail := &query.GetUserByEmail{Email: oauthUser.Result.Email}
			err = bus.Dispatch(c, userByEmail)
			user = userByEmail.Result
		}
		if err != nil {
			if errors.Cause(err) == app.ErrNotFound {
				customOAuthConfigByProvider := &query.GetCustomOAuthConfigByProvider{Provider: provider}
				err = bus.Dispatch(c, customOAuthConfigByProvider)
				customOAuthConfig := customOAuthConfigByProvider.Result
				if err != nil {
					customOAuthConfig = &models.OAuthConfig{IsTrusted: false}
				}
				if c.Tenant().IsPrivate && !customOAuthConfig.IsTrusted {
					return c.Redirect("/not-invited")
				}

				user = &models.User{
					Name:   oauthUser.Result.Name,
					Tenant: c.Tenant(),
					Email:  oauthUser.Result.Email,
					Role:   enum.RoleVisitor,
					Providers: []*models.UserProvider{
						&models.UserProvider{
							UID:  oauthUser.Result.ID,
							Name: provider,
						},
					},
				}

				if err = bus.Dispatch(c, &cmd.RegisterUser{User: user}); err != nil {
					return c.Failure(err)
				}
			} else {
				return c.Failure(err)
			}
		} else if !user.HasProvider(provider) {
			if err = bus.Dispatch(c, &cmd.RegisterUserProvider{
				UserID:       user.ID,
				ProviderName: provider,
				ProviderUID:  oauthUser.Result.ID,
			}); err != nil {
				return c.Failure(err)
			}
		}

		webutil.AddAuthUserCookie(c, user)

		return c.Redirect(redirectURL.String())
	}
}

// OAuthCallback handles the redirect back from the OAuth provider
// This callback can run on either Tenant or Login address
// If the request is for a sign in, we redirect the user to the tenant address
// If the request is for a sign up, we exchange the OAuth code and get the user profile
func OAuthCallback() web.HandlerFunc {
	return func(c *web.Context) error {
		c.Response.Header().Add("X-Robots-Tag", "noindex")

		provider := c.Param("provider")
		state := c.QueryParam("state")
		parts := strings.Split(state, "|")

		redirectURL, err := url.ParseRequestURI(parts[0])
		if err != nil {
			return c.Failure(err)
		}

		code := c.QueryParam("code")
		if code == "" {
			return c.Redirect(redirectURL.String())
		}

		//Test OAuth
		if redirectURL.Path == fmt.Sprintf("/oauth/%s/echo", provider) {
			var query = redirectURL.Query()
			query.Set("code", code)
			query.Set("identifier", parts[1])
			redirectURL.RawQuery = query.Encode()
			return c.Redirect(redirectURL.String())
		}

		//Sign up process
		if redirectURL.Path == "/signup" {
			oauthUser := &query.GetOAuthProfile{Provider: provider, Code: code}
			if err := bus.Dispatch(c, oauthUser); err != nil {
				return c.Failure(err)
			}

			claims := jwt.OAuthClaims{
				OAuthID:       oauthUser.Result.ID,
				OAuthProvider: provider,
				OAuthName:     oauthUser.Result.Name,
				OAuthEmail:    oauthUser.Result.Email,
				Metadata: jwt.Metadata{
					ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
				},
			}

			token, err := jwt.Encode(claims)
			if err != nil {
				return c.Failure(err)
			}

			var query = redirectURL.Query()
			query.Set("token", token)
			redirectURL.RawQuery = query.Encode()
			return c.Redirect(redirectURL.String())
		}

		//Sign in process
		var query = redirectURL.Query()
		query.Set("code", code)
		query.Set("redirect", redirectURL.RequestURI())
		query.Set("identifier", parts[1])
		redirectURL.RawQuery = query.Encode()
		redirectURL.Path = fmt.Sprintf("/oauth/%s/token", provider)
		return c.Redirect(redirectURL.String())
	}
}

// SignInByOAuth is responsible for redirecting the user to the OAuth authorization URL for given provider
// A cookie is stored in user's browser with a random identifier that is later used to verify the authenticity of the request
func SignInByOAuth() web.HandlerFunc {
	return func(c *web.Context) error {
		c.Response.Header().Add("X-Robots-Tag", "noindex")

		provider := c.Param("provider")
		redirect := c.QueryParam("redirect")

		if redirect == "" {
			redirect = c.BaseURL()
		}

		redirectURL, _ := url.ParseRequestURI(redirect)
		redirectURL.ResolveReference(c.Request.URL)

		if c.IsAuthenticated() && redirectURL.Path != fmt.Sprintf("/oauth/%s/echo", provider) {
			return c.Redirect(redirect)
		}

		authURL := &query.GetOAuthAuthorizationURL{
			Provider:   provider,
			Redirect:   redirect,
			Identifier: c.SessionID(),
		}
		if err := bus.Dispatch(c, authURL); err != nil {
			return c.Failure(err)
		}
		return c.Redirect(authURL.Result)
	}
}
