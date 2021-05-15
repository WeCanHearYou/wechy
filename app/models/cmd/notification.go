package cmd

import (
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/models/entities"
)

type MarkAllNotificationsAsRead struct{}

type PurgeExpiredNotifications struct {
	NumOfDeletedNotifications int
}

type MarkNotificationAsRead struct {
	ID int
}

type AddNewNotification struct {
	User   *entities.User
	Title  string
	Link   string
	PostID int

	Result *entities.Notification
}

type AddSubscriber struct {
	Post *models.Post
	User *entities.User
}

type RemoveSubscriber struct {
	Post *models.Post
	User *entities.User
}
