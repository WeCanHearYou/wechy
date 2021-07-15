package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/getfider/fider/app"

	"net/http"
	"net/http/httptest"

	"github.com/getfider/fider/app/handlers"
	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/models/query"
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/jwt"
	"github.com/getfider/fider/app/pkg/mock"
	"github.com/getfider/fider/app/pkg/web"
)

func TestSignInByEmailHandler_WithoutEmail(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	code, _ := server.
		OnTenant(mock.DemoTenant).
		ExecutePost(handlers.SignInByEmail(), "{ }")

	Expect(code).Equals(http.StatusBadRequest)
}

func TestSignInByEmailHandler_WithEmail(t *testing.T) {
	RegisterT(t)

	var saveKeyCmd *cmd.SaveVerificationKey
	bus.AddHandler(func(ctx context.Context, c *cmd.SaveVerificationKey) error {
		saveKeyCmd = c
		return nil
	})

	server := mock.NewServer()
	code, _ := server.
		OnTenant(mock.DemoTenant).
		ExecutePost(handlers.SignInByEmail(), `{ "email": "jon.snow@got.com" }`)

	Expect(code).Equals(http.StatusOK)
	Expect(saveKeyCmd.Key).HasLen(64)
	Expect(saveKeyCmd.Request.GetKind()).Equals(enum.EmailVerificationKindSignIn)
	Expect(saveKeyCmd.Request.GetEmail()).Equals("jon.snow@got.com")
	Expect(saveKeyCmd.Request.GetName()).Equals("")
}

func TestVerifySignInKeyHandler_UnknownKey(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		return app.ErrNotFound
	})

	server := mock.NewServer()
	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=unknown").
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusNotFound)
}

func TestVerifySignInKeyHandler_UsedKey(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			now := time.Now()
			q.Result = &entity.EmailVerification{
				Key:        q.Key,
				Kind:       q.Kind,
				VerifiedAt: &now,
				Email:      "jon.snow@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusGone)
}

func TestVerifySignInKeyHandler_ExpiredKey(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			now := time.Now().Add(-5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: now,
				Email:     "jon.snow@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	verified := false
	bus.AddHandler(func(ctx context.Context, c *cmd.SetKeyAsVerified) error {
		if c.Key == key {
			verified = true
		}
		return nil
	})

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusGone)
	Expect(verified).IsTrue()
}

func TestVerifySignInKeyHandler_CorrectKey_ExistingUser(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "jon.snow@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		if q.Email == mock.JonSnow.Email {
			q.Result = mock.JonSnow
			return nil
		}
		return app.ErrNotFound
	})

	verified := false
	bus.AddHandler(func(ctx context.Context, c *cmd.SetKeyAsVerified) error {
		if c.Key == key {
			verified = true
		}
		return nil
	})

	code, response := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusTemporaryRedirect)
	Expect(response.Header().Get("Location")).Equals("http://demo.test.fider.io")

	Expect(verified).IsTrue()
	ExpectFiderAuthCookie(response, mock.JonSnow)
}

func TestVerifySignInKeyHandler_CorrectKey_NewUser(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.SearchPosts) error {
		return nil
	})

	bus.AddHandler(func(ctx context.Context, q *query.CountPostPerStatus) error {
		return nil
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetAllTags) error {
		return nil
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		return app.ErrNotFound
	})

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	server := mock.NewServer()

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusOK)
}

func TestVerifySignInKeyHandler_PrivateTenant_SignInRequest_NonInviteNewUser(t *testing.T) {
	RegisterT(t)

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		return app.ErrNotFound
	})

	server := mock.NewServer()
	mock.DemoTenant.IsPrivate = true

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusForbidden)
}

func TestVerifySignInKeyHandler_PrivateTenant_SignInRequest_RegisteredUser(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	mock.DemoTenant.IsPrivate = true

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	user := &entity.User{
		Name:   "Hot Pie",
		Email:  "hot.pie@got.com",
		Tenant: mock.DemoTenant,
	}

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		if q.Email == user.Email {
			q.Result = user
			return nil
		}
		return app.ErrNotFound
	})

	verified := false
	bus.AddHandler(func(ctx context.Context, c *cmd.SetKeyAsVerified) error {
		if c.Key == key {
			verified = true
		}
		return nil
	})

	code, response := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindSignIn))

	Expect(code).Equals(http.StatusTemporaryRedirect)
	Expect(response.Header().Get("Location")).Equals("http://demo.test.fider.io")
	Expect(verified).IsTrue()
}

func TestVerifySignInKeyHandler_PrivateTenant_InviteRequest_ExistingUser(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	mock.DemoTenant.IsPrivate = true

	user := &entity.User{
		Name:   "Hot Pie",
		Email:  "hot.pie@got.com",
		Tenant: mock.DemoTenant,
	}

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindUserInvitation {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	verified := false
	bus.AddHandler(func(ctx context.Context, c *cmd.SetKeyAsVerified) error {
		if c.Key == key {
			verified = true
		}
		return nil
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		if q.Email == user.Email {
			q.Result = user
			return nil
		}
		return app.ErrNotFound
	})

	code, response := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/invite/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindUserInvitation))

	Expect(code).Equals(http.StatusTemporaryRedirect)
	Expect(response.Header().Get("Location")).Equals("http://demo.test.fider.io")
	Expect(verified).IsTrue()
}

func TestVerifySignInKeyHandler_PrivateTenant_InviteRequest_NewUser(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	mock.DemoTenant.IsPrivate = true

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindUserInvitation {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		return app.ErrNotFound
	})

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/invite/verify?k=" + key).
		Execute(handlers.VerifySignInKey(enum.EmailVerificationKindUserInvitation))

	Expect(code).Equals(http.StatusOK)
}

func TestVerifySignUpKeyHandler_PendingTenant(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	mock.DemoTenant.Status = enum.TenantPending

	var newUser *entity.User
	bus.AddHandler(func(ctx context.Context, c *cmd.RegisterUser) error {
		newUser = c.User
		return nil
	})

	activated := false
	bus.AddHandler(func(ctx context.Context, c *cmd.ActivateTenant) error {
		activated = true
		return nil
	})

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignUp {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Name:      "Hot Pie",
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	verified := false
	bus.AddHandler(func(ctx context.Context, c *cmd.SetKeyAsVerified) error {
		if c.Key == key {
			verified = true
		}
		return nil
	})

	code, response := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signup/verify?k=" + key).
		Execute(handlers.VerifySignUpKey())

	Expect(code).Equals(http.StatusTemporaryRedirect)
	Expect(response.Header().Get("Location")).Equals("http://demo.test.fider.io")
	Expect(newUser.Name).Equals("Hot Pie")
	Expect(newUser.Email).Equals("hot.pie@got.com")
	Expect(activated).IsTrue()
	Expect(verified).IsTrue()
}

func TestCompleteSignInProfileHandler_UnknownKey(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/complete").
		ExecutePost(handlers.CompleteSignInProfile(), `{ }`)

	Expect(code).Equals(http.StatusBadRequest)
}

func TestCompleteSignInProfileHandler_ExistingUser_CorrectKey(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		if q.Email == mock.JonSnow.Email {
			q.Result = mock.JonSnow
			return nil
		}
		return app.ErrNotFound
	})

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     mock.JonSnow.Email,
			}
			return nil
		}
		return app.ErrNotFound
	})

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/complete").
		ExecutePost(handlers.CompleteSignInProfile(), `{ "name": "Hot Pie", "key": "`+key+`" }`)

	Expect(code).Equals(http.StatusOK)
}

func TestCompleteSignInProfileHandler_CorrectKey(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()

	var newUser *entity.User
	bus.AddHandler(func(ctx context.Context, c *cmd.RegisterUser) error {
		newUser = c.User
		return nil
	})

	bus.AddHandler(func(ctx context.Context, q *query.GetUserByEmail) error {
		return app.ErrNotFound
	})

	key := "1234567890"
	bus.AddHandler(func(ctx context.Context, q *query.GetVerificationByKey) error {
		if q.Key == key && q.Kind == enum.EmailVerificationKindSignIn {
			expiresAt := time.Now().Add(5 * time.Minute)
			q.Result = &entity.EmailVerification{
				Key:       q.Key,
				Kind:      q.Kind,
				ExpiresAt: expiresAt,
				Email:     "hot.pie@got.com",
			}
			return nil
		}
		return app.ErrNotFound
	})

	verified := false
	bus.AddHandler(func(ctx context.Context, c *cmd.SetKeyAsVerified) error {
		if c.Key == key {
			verified = true
		}
		return nil
	})

	code, response := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin/complete").
		ExecutePost(handlers.CompleteSignInProfile(), `{ "name": "Hot Pie", "key": "`+key+`" }`)
	Expect(code).Equals(http.StatusOK)

	Expect(newUser.Name).Equals("Hot Pie")
	Expect(newUser.Email).Equals("hot.pie@got.com")
	Expect(verified).IsTrue()

	ExpectFiderAuthCookie(response, newUser)
}

func TestSignInPageHandler_AuthenticatedUser(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	code, response := server.
		OnTenant(mock.DemoTenant).
		AsUser(mock.AryaStark).
		WithURL("http://demo.test.fider.io/signin").
		Execute(handlers.SignInPage())

	Expect(code).Equals(http.StatusTemporaryRedirect)
	Expect(response.Header().Get("Location")).Equals("http://demo.test.fider.io")
}

func TestSignInPageHandler_NonPrivateTenant(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	code, response := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin").
		Execute(handlers.SignInPage())

	Expect(code).Equals(http.StatusTemporaryRedirect)
	Expect(response.Header().Get("Location")).Equals("http://demo.test.fider.io")
}

func TestSignInPageHandler_PrivateTenant_UnauthenticatedUser(t *testing.T) {
	RegisterT(t)

	server := mock.NewServer()
	mock.DemoTenant.IsPrivate = true

	code, _ := server.
		OnTenant(mock.DemoTenant).
		WithURL("http://demo.test.fider.io/signin").
		Execute(handlers.SignInPage())

	Expect(code).Equals(http.StatusOK)
}

func ExpectFiderAuthCookie(response *httptest.ResponseRecorder, expected *entity.User) {
	cookies := response.Header()["Set-Cookie"]
	if expected == nil {
		for _, c := range cookies {
			cookie := web.ParseCookie(c)
			Expect(cookie.Name).NotEquals(web.CookieAuthName)
		}
	} else {
		for _, c := range cookies {
			cookie := web.ParseCookie(c)
			if cookie.Name == web.CookieAuthName {
				Expect(cookie.Name).Equals(web.CookieAuthName)
				ExpectFiderToken(cookie.Value, expected)
				Expect(cookie.HttpOnly).IsTrue()
				Expect(cookie.Path).Equals("/")
				Expect(cookie.Expires).TemporarilySimilar(time.Now().Add(365*24*time.Hour), 5*time.Second)
				return
			}
		}
		panic("Cookie not found...")
	}
}

func ExpectFiderToken(token string, expected *entity.User) {
	user, err := jwt.DecodeFiderClaims(token)
	Expect(err).IsNil()
	Expect(user.UserID).Equals(expected.ID)
	Expect(user.UserName).Equals(expected.Name)
	Expect(user.UserEmail).Equals(expected.Email)
}
