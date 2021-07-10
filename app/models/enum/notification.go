package enum

import (
	"strconv"
)

//NotificationChannel represents the medium that the notification is sent
type NotificationChannel int

var (
	//NotificationChannelWeb is a in-app notification
	NotificationChannelWeb NotificationChannel = 1
	//NotificationChannelEmail is an email notification
	NotificationChannelEmail NotificationChannel = 2
)

//NotificationEvent represents all possible notification events
type NotificationEvent struct {
	UserSettingsKeyName           string
	DefaultSettingValue           string
	RequiresSubscriptionUserRoles []Role
	DefaultEnabledUserRoles       []Role
	Validate                      func(string) bool
}

func notificationEventValidation(v string) bool {
	return v == "0" || v == "1" || v == "2" || v == "3"
}

var (
	//NotificationEventNewPost is triggered when a new post is posted
	NotificationEventNewPost = NotificationEvent{
		UserSettingsKeyName:           "event_notification_new_post",
		DefaultSettingValue:           strconv.Itoa(int(NotificationChannelWeb | NotificationChannelEmail)),
		RequiresSubscriptionUserRoles: []Role{},
		DefaultEnabledUserRoles: []Role{
			RoleAdministrator,
			RoleCollaborator,
		},
		Validate: notificationEventValidation,
	}
	//NotificationEventNewComment is triggered when a new comment is posted
	NotificationEventNewComment = NotificationEvent{
		UserSettingsKeyName: "event_notification_new_comment",
		DefaultSettingValue: strconv.Itoa(int(NotificationChannelWeb | NotificationChannelEmail)),
		RequiresSubscriptionUserRoles: []Role{
			RoleVisitor,
		},
		DefaultEnabledUserRoles: []Role{
			RoleAdministrator,
			RoleCollaborator,
			RoleVisitor,
		},
		Validate: notificationEventValidation,
	}
	//NotificationEventChangeStatus is triggered when a new post has its status changed
	NotificationEventChangeStatus = NotificationEvent{
		UserSettingsKeyName: "event_notification_change_status",
		DefaultSettingValue: strconv.Itoa(int(NotificationChannelWeb | NotificationChannelEmail)),
		RequiresSubscriptionUserRoles: []Role{
			RoleVisitor,
		},
		DefaultEnabledUserRoles: []Role{
			RoleAdministrator,
			RoleCollaborator,
			RoleVisitor,
		},
		Validate: notificationEventValidation,
	}
	//AllNotificationEvents contains all possible notification events
	AllNotificationEvents = []NotificationEvent{
		NotificationEventNewPost,
		NotificationEventNewComment,
		NotificationEventChangeStatus,
	}
)
