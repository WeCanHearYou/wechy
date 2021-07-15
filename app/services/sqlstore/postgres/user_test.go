package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/models/query"

	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/pkg/bus"

	"github.com/getfider/fider/app"
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/pkg/errors"
)

func TestUserStorage_GetByID(t *testing.T) {
	ctx := SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	userByID := &query.GetUserByID{UserID: 1}
	err := bus.Dispatch(ctx, userByID)
	Expect(err).IsNil()
	Expect(userByID.Result.ID).Equals(int(1))
	Expect(userByID.Result.Tenant.ID).Equals(1)
	Expect(userByID.Result.Name).Equals("Jon Snow")
	Expect(userByID.Result.Email).Equals("jon.snow@got.com")
	Expect(userByID.Result.AvatarURL).Equals("http://cdn.test.fider.io/static/avatars/gravatar/1/Jon%20Snow")
	Expect(userByID.Result.Providers).HasLen(1)
	Expect(userByID.Result.Providers[0].UID).Equals("FB1234")
	Expect(userByID.Result.Providers[0].Name).Equals("facebook")
}

func TestUserStorage_GetByEmail_Error(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	userByEmail := &query.GetUserByEmail{Email: "unknown@got.com"}
	err := bus.Dispatch(demoTenantCtx, userByEmail)
	Expect(err).IsNotNil()
	Expect(userByEmail.Result).IsNil()
}

func TestUserStorage_GetByEmail(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	userByEmail := &query.GetUserByEmail{Email: "jon.snow@got.com"}
	userByUpperCaseEmail := &query.GetUserByEmail{Email: "JON.SNOW@got.com"}
	err := bus.Dispatch(demoTenantCtx, userByEmail, userByUpperCaseEmail)
	Expect(err).IsNil()

	Expect(userByEmail.Result.ID).Equals(int(1))
	Expect(userByEmail.Result.Tenant.ID).Equals(1)
	Expect(userByEmail.Result.Name).Equals("Jon Snow")
	Expect(userByEmail.Result.Email).Equals("jon.snow@got.com")
	Expect(userByEmail.Result.Providers).HasLen(1)
	Expect(userByEmail.Result.Providers[0].UID).Equals("FB1234")
	Expect(userByEmail.Result.Providers[0].Name).Equals("facebook")

	Expect(userByUpperCaseEmail.Result.ID).Equals(int(1))
}

func TestUserStorage_GetByProvider(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	getUser := &query.GetUserByProvider{Provider: "facebook", UID: "FB1234"}
	err := bus.Dispatch(demoTenantCtx, getUser)
	Expect(err).IsNil()

	Expect(getUser.Result.ID).Equals(int(1))
	Expect(getUser.Result.Tenant.ID).Equals(1)
	Expect(getUser.Result.Name).Equals("Jon Snow")
	Expect(getUser.Result.Email).Equals("jon.snow@got.com")
	Expect(getUser.Result.Providers).HasLen(1)
	Expect(getUser.Result.Providers[0].UID).Equals("FB1234")
	Expect(getUser.Result.Providers[0].Name).Equals("facebook")
}

func TestUserStorage_GetByProvider_WrongTenant(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	getUser := &query.GetUserByProvider{Provider: "facebook", UID: "FB1234"}
	err := bus.Dispatch(avengersTenantCtx, getUser)
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
	Expect(getUser.Result).IsNil()
}

func TestUserStorage_GetByEmail_WrongTenant(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	getUser := &query.GetUserByEmail{Email: "jon.snow@got.com"}
	err := bus.Dispatch(avengersTenantCtx, getUser)
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
	Expect(getUser.Result).IsNil()
}

func TestUserStorage_Register(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	user := &entity.User{
		Name:  "Rob Stark",
		Email: "rob.stark@got.com",
		Role:  enum.RoleCollaborator,
		Providers: []*entity.UserProvider{
			{
				UID:  "123123123",
				Name: app.FacebookProvider,
			},
		},
	}
	err := bus.Dispatch(demoTenantCtx, &cmd.RegisterUser{User: user})
	Expect(err).IsNil()

	getUser := &query.GetUserByEmail{Email: "rob.stark@got.com"}
	err = bus.Dispatch(demoTenantCtx, getUser)
	Expect(err).IsNil()

	Expect(getUser.Result.ID).Equals(int(6))
	Expect(getUser.Result.Role).Equals(enum.RoleCollaborator)
	Expect(getUser.Result.Name).Equals("Rob Stark")
	Expect(getUser.Result.Email).Equals("rob.stark@got.com")
	Expect(getUser.Result.Status).Equals(enum.UserActive)
}

func TestUserStorage_Register_WhiteSpaceEmail(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	user := &entity.User{
		Name:  "Rob Stark",
		Email: "   ",
		Role:  enum.RoleCollaborator,
	}
	err := bus.Dispatch(demoTenantCtx, &cmd.RegisterUser{User: user})
	Expect(err).IsNil()

	getUser := &query.GetUserByID{UserID: user.ID}
	err = bus.Dispatch(demoTenantCtx, getUser)
	Expect(err).IsNil()

	Expect(getUser.Result.Role).Equals(enum.RoleCollaborator)
	Expect(getUser.Result.Name).Equals("Rob Stark")
	Expect(getUser.Result.Email).Equals("")
	Expect(getUser.Result.Status).Equals(enum.UserActive)
}

func TestUserStorage_Register_MultipleProviders(t *testing.T) {
	ctx := SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	var tenantID int
	err := trx.Get(&tenantID, `
		INSERT INTO tenants (name, subdomain, created_at, status, is_private, custom_css, logo_bkey, locale, is_allowing_email_auth) 
		VALUES ('My Domain Inc.','mydomain', now(), 1, false, '', '', 'en', true)
		RETURNING id
	`)
	Expect(err).IsNil()

	newTenantCtx := context.WithValue(ctx, app.TenantCtxKey, &entity.Tenant{ID: tenantID})

	user := &entity.User{
		Name:  "Jon Snow",
		Email: "jon.snow@got.com",
		Role:  enum.RoleCollaborator,
		Providers: []*entity.UserProvider{
			{
				UID:  "123123123",
				Name: app.FacebookProvider,
			},
			{
				UID:  "456456456",
				Name: app.GoogleProvider,
			},
		},
	}

	err = bus.Dispatch(newTenantCtx, &cmd.RegisterUser{User: user})
	Expect(err).IsNil()
	Expect(user.ID).NotEquals(0)
	Expect(user.Name).Equals("Jon Snow")
	Expect(user.Email).Equals("jon.snow@got.com")
	Expect(user.Status).Equals(enum.UserActive)
}

func TestUserStorage_RegisterProvider(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	err := bus.Dispatch(demoTenantCtx, &cmd.RegisterUserProvider{
		UserID:       1,
		ProviderName: "google",
		ProviderUID:  "GO1234",
	})
	Expect(err).IsNil()

	getUser := &query.GetUserByID{UserID: 1}
	err = bus.Dispatch(demoTenantCtx, getUser)
	Expect(err).IsNil()

	Expect(getUser.Result.ID).Equals(int(1))
	Expect(getUser.Result.Name).Equals("Jon Snow")
	Expect(getUser.Result.Email).Equals("jon.snow@got.com")
	Expect(getUser.Result.Tenant.ID).Equals(1)
	Expect(getUser.Result.Status).Equals(enum.UserActive)

	Expect(getUser.Result.Providers).HasLen(2)
	Expect(getUser.Result.Providers[0].UID).Equals("FB1234")
	Expect(getUser.Result.Providers[0].Name).Equals("facebook")
	Expect(getUser.Result.Providers[1].UID).Equals("GO1234")
	Expect(getUser.Result.Providers[1].Name).Equals("google")
}

func TestUserStorage_UpdateSettings(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	err := bus.Dispatch(jonSnowCtx, &cmd.UpdateCurrentUser{
		Name: "Jon Stark",
		Avatar: &dto.ImageUpload{
			BlobKey: "jon.png",
		},
	})
	Expect(err).IsNil()

	getUser := &query.GetUserByEmail{Email: "jon.snow@got.com"}
	err = bus.Dispatch(jonSnowCtx, getUser)
	Expect(err).IsNil()
	Expect(getUser.Result.Name).Equals("Jon Stark")
	Expect(getUser.Result.AvatarBlobKey).Equals("jon.png")
}

func TestUserStorage_ChangeRole(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	err := bus.Dispatch(demoTenantCtx, &cmd.ChangeUserRole{
		UserID: jonSnow.ID,
		Role:   enum.RoleVisitor,
	})
	Expect(err).IsNil()

	getUser := &query.GetUserByEmail{Email: "jon.snow@got.com"}
	err = bus.Dispatch(demoTenantCtx, getUser)
	Expect(err).IsNil()
	Expect(getUser.Result.Role).Equals(enum.RoleVisitor)
}

func TestUserStorage_ChangeEmail(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	err := bus.Dispatch(demoTenantCtx, &cmd.ChangeUserEmail{
		UserID: jonSnow.ID,
		Email:  "jon.stark@got.com",
	})
	Expect(err).IsNil()

	getByNewEmail := &query.GetUserByEmail{Email: "jon.stark@got.com"}
	err = bus.Dispatch(demoTenantCtx, getByNewEmail)
	Expect(err).IsNil()
	Expect(getByNewEmail.Result.Email).Equals("jon.stark@got.com")

	getByOldEmail := &query.GetUserByEmail{Email: "jon.snow@got.com"}
	err = bus.Dispatch(demoTenantCtx, getByOldEmail)
	Expect(getByOldEmail.Result).IsNil()
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
}

func TestUserStorage_GetAll(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	allUsers := &query.GetAllUsers{}
	err := bus.Dispatch(demoTenantCtx, allUsers)
	Expect(err).IsNil()

	Expect(allUsers.Result).HasLen(3)
	Expect(allUsers.Result[0].Name).Equals("Jon Snow")
	Expect(allUsers.Result[0].Status).Equals(enum.UserActive)
	Expect(allUsers.Result[1].Name).Equals("Arya Stark")
	Expect(allUsers.Result[1].Status).Equals(enum.UserActive)
	Expect(allUsers.Result[2].Name).Equals("Sansa Stark")
	Expect(allUsers.Result[2].Status).Equals(enum.UserActive)
}

func TestUserStorage_DefaultUserSettings(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	getSettings := &query.GetCurrentUserSettings{}

	err := bus.Dispatch(jonSnowCtx, getSettings)
	Expect(err).IsNil()

	Expect(getSettings.Result).Equals(map[string]string{
		enum.NotificationEventNewPost.UserSettingsKeyName:      enum.NotificationEventNewPost.DefaultSettingValue,
		enum.NotificationEventNewComment.UserSettingsKeyName:   enum.NotificationEventNewComment.DefaultSettingValue,
		enum.NotificationEventChangeStatus.UserSettingsKeyName: enum.NotificationEventChangeStatus.DefaultSettingValue,
	})

	err = bus.Dispatch(aryaStarkCtx, getSettings)
	Expect(err).IsNil()
	Expect(getSettings.Result).Equals(map[string]string{
		enum.NotificationEventChangeStatus.UserSettingsKeyName: enum.NotificationEventChangeStatus.DefaultSettingValue,
	})
}

func TestUserStorage_SaveGetUserSettings(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	disableAll := map[string]string{
		enum.NotificationEventNewPost.UserSettingsKeyName:      "0",
		enum.NotificationEventChangeStatus.UserSettingsKeyName: "1",
	}

	err := bus.Dispatch(aryaStarkCtx, &cmd.UpdateCurrentUserSettings{Settings: disableAll})
	Expect(err).IsNil()

	firstSettings := &query.GetCurrentUserSettings{}
	err = bus.Dispatch(aryaStarkCtx, firstSettings)
	Expect(err).IsNil()

	Expect(firstSettings.Result).Equals(map[string]string{
		enum.NotificationEventNewPost.UserSettingsKeyName:      "0",
		enum.NotificationEventChangeStatus.UserSettingsKeyName: "1",
	})

	err = bus.Dispatch(aryaStarkCtx, &cmd.UpdateCurrentUserSettings{Settings: nil})
	Expect(err).IsNil()

	secondSettings := &query.GetCurrentUserSettings{}
	err = bus.Dispatch(aryaStarkCtx, secondSettings)
	Expect(err).IsNil()

	Expect(secondSettings.Result).Equals(firstSettings.Result)
}

func TestUserStorage_Delete(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	err := bus.Dispatch(jonSnowCtx, &cmd.DeleteCurrentUser{})
	Expect(err).IsNil()

	getByEmail := &query.GetUserByEmail{Email: "jon.snow@got.com"}
	err = bus.Dispatch(jonSnowCtx, getByEmail)
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
	Expect(getByEmail.Result).IsNil()

	getByID := &query.GetUserByID{UserID: jonSnow.ID}
	err = bus.Dispatch(jonSnowCtx, getByID)
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
	Expect(getByID.Result).IsNil()
}

func TestUserStorage_APIKey(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	regenerateAPIKey := &cmd.RegenerateAPIKey{}
	err := bus.Dispatch(jonSnowCtx, regenerateAPIKey)
	Expect(err).IsNil()
	Expect(regenerateAPIKey.Result).HasLen(64)

	firstKey := regenerateAPIKey.Result

	getByKey := &query.GetUserByAPIKey{APIKey: firstKey}
	err = bus.Dispatch(jonSnowCtx, getByKey)
	Expect(getByKey.Result).Equals(jonSnow)
	Expect(err).IsNil()

	//try to get by uppercase key
	getByKey = &query.GetUserByAPIKey{APIKey: strings.ToUpper(firstKey)}
	err = bus.Dispatch(jonSnowCtx, getByKey)
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
	Expect(getByKey.Result).IsNil()

	//regenerate and try to get again using old key
	err = bus.Dispatch(jonSnowCtx, regenerateAPIKey)
	Expect(err).IsNil()

	getByKey = &query.GetUserByAPIKey{APIKey: firstKey}
	err = bus.Dispatch(jonSnowCtx, getByKey)
	Expect(getByKey.Result).IsNil()
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)

	//try to get by some unknown key
	getByKey = &query.GetUserByAPIKey{APIKey: "SOME-INVALID-KEY"}
	err = bus.Dispatch(jonSnowCtx, getByKey)
	Expect(getByKey.Result).IsNil()
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
}

func TestUserStorage_BlockUser(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	userID := 1
	getUser := &query.GetUserByID{UserID: userID}

	err := bus.Dispatch(demoTenantCtx, getUser)
	Expect(err).IsNil()
	Expect(getUser.Result.Status).Equals(enum.UserActive)

	err = bus.Dispatch(demoTenantCtx, &cmd.BlockUser{UserID: userID}, getUser)
	Expect(err).IsNil()
	Expect(getUser.Result.Status).Equals(enum.UserBlocked)

	err = bus.Dispatch(demoTenantCtx, &cmd.UnblockUser{UserID: userID}, getUser)
	Expect(err).IsNil()
	Expect(getUser.Result.Status).Equals(enum.UserActive)
}
