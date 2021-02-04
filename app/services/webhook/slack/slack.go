package slack

import (
	"context"
	"encoding/json"
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
	return "Slack"
}

func (s Service) Category() string {
	return "Webhook"
}

func (s Service) Enabled() bool {
	return env.Config.Webhook.Slack.URL != ""
}

func (s Service) Init() {
	notifierClient = &webhook.Client{
		Notifier: "Slack",
		URL:      env.Config.Webhook.Slack.URL,
	}
	bus.AddListener(sendWebhookNotification)
}

func sendWebhookNotification(ctx context.Context, c *cmd.SendWebhookNotification) {
	message := RenderMessage(c.Type, c.Title, c.Content, c.Link)
	raw, err := json.Marshal(message)
	if err != nil {
		log.Errorf(ctx, "marshal Slack message failed: @{Error}.", dto.Props{
			"Error": err,
		})
		return
	}

	err = notifierClient.Notify(ctx, raw)
	if err != nil {
		log.Errorf(ctx, "failed to send a notification through Slack webhook: @{Error}.", dto.Props{
			"Error": err,
		})
	}
}
