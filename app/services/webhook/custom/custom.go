package custom

import (
	"context"
	"encoding/json"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/log"
	"github.com/getfider/fider/app/services/webhook"
)

var notifierClient *webhook.Client

func init() {
	bus.Register(Service{})
}

type Service struct{}

func (s Service) Name() string {
	return "Custom"
}

func (s Service) Category() string {
	return "Webhook"
}

func (s Service) Enabled() bool {
	return env.Config.Webhook.Custom.URL != ""
}

func (s Service) Init() {
	notifierClient = &webhook.Client{
		Notifier: "Custom",
		URL:      env.Config.Webhook.Custom.URL,
	}
	bus.AddListener(sendWebhookNotification)
}

type Message struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Link        string      `json:"link"`
	User        models.User `json:"user"`
}

func sendWebhookNotification(ctx context.Context, c *cmd.SendWebhookNotification) {
	message := Message{
		Title:       c.Title,
		Description: c.Content,
		Link:        c.Link,
		User:        c.User,
	}

	raw, err := json.Marshal(message)
	if err != nil {
		log.Errorf(ctx, "marshal Custom message failed: @{Error}.", dto.Props{
			"Error": err,
		})
		return
	}

	err = notifierClient.Notify(ctx, raw)
	if err != nil {
		log.Errorf(ctx, "failed to send a notification through Custom webhook: @{Error}.", dto.Props{
			"Error": err,
		})
	}
}
