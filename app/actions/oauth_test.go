package actions_test

import (
	"context"
	"github.com/getfider/fider/app/models/dto"
	"testing"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/actions"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/models/query"
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/rand"
)

func TestCreateEditOAuthConfig_InvalidInput(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.ListActiveOAuthProviders) error {
		q.Result = []*dto.OAuthProviderOption{}
		return nil
	})

	testCases := []struct {
		expected []string
		action   *actions.CreateEditOAuthConfig
	}{
		{
			expected: []string{"displayName", "status", "tokenURL", "clientID", "clientSecret", "scope", "authorizeURL", "tokenURL", "jsonUserIDPath"},
			action:   &actions.CreateEditOAuthConfig{},
		},
		{
			expected: []string{"displayName", "status", "tokenURL", "clientID", "clientSecret", "scope", "authorizeURL", "tokenURL", "profileURL", "jsonUserIDPath", "jsonUserNamePath", "jsonUserEmailPath"},
			action: &actions.CreateEditOAuthConfig{
				DisplayName:       rand.String(51),
				ClientID:          rand.String(101),
				Status:            0,
				ClientSecret:      rand.String(501),
				AuthorizeURL:      rand.String(301),
				TokenURL:          rand.String(301),
				Scope:             rand.String(101),
				ProfileURL:        rand.String(301),
				JSONUserIDPath:    rand.String(101),
				JSONUserNamePath:  rand.String(101),
				JSONUserEmailPath: rand.String(101),
			},
		},
	}
	ctx := context.WithValue(context.Background(), app.TenantCtxKey, &entity.Tenant{
		IsAllowingEmailAuth: true,
	})

	for _, testCase := range testCases {
		result := testCase.action.Validate(ctx, nil)
		ExpectFailed(result, testCase.expected...)
	}
}

func TestCreateEditOAuthConfig_DefaultValues(t *testing.T) {
	RegisterT(t)

	action := actions.NewCreateEditOAuthConfig()
	Expect(action.Logo.BlobKey).Equals("")
}

func TestCreateEditOAuthConfig_AddNew_ValidInput(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.ListActiveOAuthProviders) error {
		q.Result = []*dto.OAuthProviderOption{}
		return nil
	})

	action := &actions.CreateEditOAuthConfig{
		DisplayName:       "My Provider",
		Status:            enum.OAuthConfigEnabled,
		ClientID:          "823187ahjjfdha8fds7yfdashfjkdsa",
		ClientSecret:      "jijads78d76cn347768x3t4668q275@ˆ&Tnycasdgsacuyhij",
		AuthorizeURL:      "http://provider/oauth/authorize",
		TokenURL:          "http://provider/oauth/token",
		Scope:             "profile email",
		ProfileURL:        "http://provider/profile/me",
		JSONUserIDPath:    "user.id",
		JSONUserNamePath:  "user.name",
		JSONUserEmailPath: "user.email",
	}
	ctx := context.WithValue(context.Background(), app.TenantCtxKey, &entity.Tenant{
		IsAllowingEmailAuth: true,
	})

	result := action.Validate(ctx, nil)
	ExpectSuccess(result)
	Expect(action.ID).Equals(0)
	Expect(action.Provider).HasLen(11)
	Expect(string(action.Provider[0])).Equals("_")
}

func TestCreateEditOAuthConfig_EditExisting_NewSecret(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.GetCustomOAuthConfigByProvider) error {
		if q.Provider == "_NAME" {
			q.Result = &entity.OAuthConfig{
				ID:          4,
				Provider:    q.Provider,
				LogoBlobKey: "hello-world.png",
			}
			return nil
		}
		return app.ErrNotFound
	})

	bus.AddHandler(func(ctx context.Context, q *query.ListActiveOAuthProviders) error {
		q.Result = []*dto.OAuthProviderOption{}
		return nil
	})

	action := actions.NewCreateEditOAuthConfig()
	action.Provider = "_NAME"
	action.DisplayName = "My Provider"
	action.Status = enum.OAuthConfigDisabled
	action.ClientID = "823187ahjjfdha8fds7yfdashfjkdsa"
	action.ClientSecret = "jijads78d76cn347768x3t4668q275@ˆ&Tnycasdgsacuyhij"
	action.AuthorizeURL = "http://provider/oauth/authorize"
	action.TokenURL = "http://provider/oauth/token"
	action.Scope = "profile email"
	action.ProfileURL = "http://provider/profile/me"
	action.JSONUserIDPath = "user.id"
	action.JSONUserNamePath = "user.name"
	action.JSONUserEmailPath = "user.email"
	ctx := context.WithValue(context.Background(), app.TenantCtxKey, &entity.Tenant{
		IsAllowingEmailAuth: true,
	})

	result := action.Validate(ctx, nil)
	ExpectSuccess(result)
	Expect(action.ID).Equals(4)
	Expect(action.Logo.BlobKey).Equals("hello-world.png")
	Expect(action.ClientSecret).Equals("jijads78d76cn347768x3t4668q275@ˆ&Tnycasdgsacuyhij")
}

func TestCreateEditOAuthConfig_EditExisting_OmitSecret(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.GetCustomOAuthConfigByProvider) error {
		if q.Provider == "_NAME2" {
			q.Result = &entity.OAuthConfig{
				ID:           5,
				Provider:     q.Provider,
				DisplayName:  "My Provider",
				ClientSecret: "MY_OLD_SECRET",
			}
			return nil
		}
		return app.ErrNotFound
	})

	bus.AddHandler(func(ctx context.Context, q *query.ListActiveOAuthProviders) error {
		q.Result = []*dto.OAuthProviderOption{}
		return nil
	})

	action := actions.NewCreateEditOAuthConfig()
	action.Provider = "_NAME2"
	action.DisplayName = "My Provider"
	action.Status = enum.OAuthConfigDisabled
	action.ClientID = "823187ahjjfdha8fds7yfdashfjkdsa"
	action.AuthorizeURL = "http://provider/oauth/authorize"
	action.TokenURL = "http://provider/oauth/token"
	action.Scope = "profile email"
	action.ProfileURL = "http://provider/profile/me"
	action.JSONUserIDPath = "user.id"
	action.JSONUserNamePath = "user.name"
	action.JSONUserEmailPath = "user.email"
	ctx := context.WithValue(context.Background(), app.TenantCtxKey, &entity.Tenant{
		IsAllowingEmailAuth: true,
	})

	result := action.Validate(ctx, nil)
	ExpectSuccess(result)
	Expect(action.ID).Equals(5)
	Expect(action.ClientSecret).Equals("MY_OLD_SECRET")
}

func TestCreateEditOAuthConfig_EditNonExisting(t *testing.T) {
	RegisterT(t)

	bus.AddHandler(func(ctx context.Context, q *query.GetCustomOAuthConfigByProvider) error {
		return app.ErrNotFound
	})

	bus.AddHandler(func(ctx context.Context, q *query.ListActiveOAuthProviders) error {
		q.Result = []*dto.OAuthProviderOption{}
		return nil
	})

	action := actions.NewCreateEditOAuthConfig()
	action.Provider = "_MY_NEW_PROVIDER"
	action.DisplayName = "My Provider"
	action.Status = enum.OAuthConfigDisabled
	action.ClientID = "823187ahjjfdha8fds7yfdashfjkdsa"
	action.AuthorizeURL = "http://provider/oauth/authorize"
	action.TokenURL = "http://provider/oauth/token"
	action.Scope = "profile email"
	action.ProfileURL = "http://provider/profile/me"
	action.JSONUserIDPath = "user.id"
	action.JSONUserNamePath = "user.name"
	action.JSONUserEmailPath = "user.email"
	ctx := context.WithValue(context.Background(), app.TenantCtxKey, &entity.Tenant{
		IsAllowingEmailAuth: true,
	})

	result := action.Validate(ctx, nil)
	Expect(result.Err).Equals(app.ErrNotFound)
	Expect(result.Ok).IsFalse()
}
