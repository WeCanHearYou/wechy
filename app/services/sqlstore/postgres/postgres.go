package postgres

import (
	"context"

	"github.com/getfider/fider/app"

	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/dbx"
)

func init() {
	bus.Register(Service{})
}

type Service struct{}

func (s Service) Name() string {
	return "PostgreSQL"
}

func (s Service) Category() string {
	return "sqlstore"
}

func (s Service) Enabled() bool {
	return true
}

func (s Service) Init() {
	bus.AddHandler(storeEvent)

	bus.AddListener(purgeExpiredNotifications)

	bus.AddHandler(markAllNotificationsAsRead)
	bus.AddHandler(markNotificationAsRead)
	bus.AddHandler(countUnreadNotifications)
	bus.AddHandler(getNotificationByID)
	bus.AddHandler(getActiveNotifications)
	bus.AddHandler(addNewNotification)
	bus.AddHandler(addSubscriber)
	bus.AddHandler(removeSubscriber)
	bus.AddHandler(getActiveSubscribers)

	bus.AddHandler(getTagBySlug)
	bus.AddHandler(getAssignedTags)
	bus.AddHandler(getAllTags)
	bus.AddHandler(addNewTag)
	bus.AddHandler(updateTag)
	bus.AddHandler(deleteTag)
	bus.AddHandler(assignTag)
	bus.AddHandler(unassignTag)

	bus.AddHandler(addVote)
	bus.AddHandler(removeVote)
	bus.AddHandler(listPostVotes)

	bus.AddHandler(addNewPost)
	bus.AddHandler(updatePost)
	bus.AddHandler(getPostByID)
	bus.AddHandler(getPostBySlug)
	bus.AddHandler(getPostByNumber)
	bus.AddHandler(searchPosts)
	bus.AddHandler(getAllPosts)
	bus.AddHandler(countPostPerStatus)
	bus.AddHandler(markPostAsDuplicate)
	bus.AddHandler(setPostResponse)
	bus.AddHandler(postIsReferenced)

	bus.AddHandler(setAttachments)
	bus.AddHandler(getAttachments)
	bus.AddHandler(uploadImage)
	bus.AddHandler(uploadImages)

	bus.AddHandler(addNewComment)
	bus.AddHandler(updateComment)
	bus.AddHandler(deleteComment)
	bus.AddHandler(getCommentByID)
	bus.AddHandler(getCommentsByPost)

	bus.AddHandler(countUsers)
	bus.AddHandler(blockUser)
	bus.AddHandler(unblockUser)
	bus.AddHandler(regenerateAPIKey)
	bus.AddHandler(userSubscribedTo)
	bus.AddHandler(deleteCurrentUser)
	bus.AddHandler(changeUserEmail)
	bus.AddHandler(changeUserRole)
	bus.AddHandler(updateCurrentUserSettings)
	bus.AddHandler(getCurrentUserSettings)
	bus.AddHandler(registerUser)
	bus.AddHandler(registerUserProvider)
	bus.AddHandler(updateCurrentUser)
	bus.AddHandler(getUserByAPIKey)
	bus.AddHandler(getUserByEmail)
	bus.AddHandler(getUserByID)
	bus.AddHandler(getUserByProvider)
	bus.AddHandler(getAllUsers)

	bus.AddHandler(createTenant)
	bus.AddHandler(getFirstTenant)
	bus.AddHandler(getTenantByDomain)
	bus.AddHandler(activateTenant)
	bus.AddHandler(isSubdomainAvailable)
	bus.AddHandler(isCNAMEAvailable)
	bus.AddHandler(updateTenantSettings)
	bus.AddHandler(updateTenantPrivacySettings)
	bus.AddHandler(updateTenantAllowingEmailAuthSettings)
	bus.AddHandler(updateTenantAdvancedSettings)

	bus.AddHandler(getVerificationByKey)
	bus.AddHandler(saveVerificationKey)
	bus.AddHandler(setKeyAsVerified)

	bus.AddHandler(listCustomOAuthConfig)
	bus.AddHandler(getCustomOAuthConfigByProvider)
	bus.AddHandler(saveCustomOAuthConfig)
}

type SqlHandler func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error

func using(ctx context.Context, handler SqlHandler) error {
	trx := ctx.Value(app.TransactionCtxKey).(*dbx.Trx)
	tenant, _ := ctx.Value(app.TenantCtxKey).(*entity.Tenant)
	user, _ := ctx.Value(app.UserCtxKey).(*entity.User)
	return handler(trx, tenant, user)
}
