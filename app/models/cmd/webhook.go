package cmd

import "github.com/getfider/fider/app/models"

type SendWebhookNotification struct {
	Type    string
	Title   string
	Link    string
	Content string
	User    models.User
}
