package slack

import (
	"fmt"
	"github.com/slack-go/slack"
)

// RenderMessage returns a Slack Message from fields given
func RenderMessage(notificationType string, title string, description string, link string) slack.WebhookMessage {
	title = fmt.Sprintf("%s *%s*", TitleEnrichWithEmojiByNotificationType(notificationType), title)

	return slack.WebhookMessage{
		Blocks: &slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", title, false, false),
					nil,
					nil,
				),
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", description, false, false),
					nil,
					&slack.Accessory{
						ButtonElement: &slack.ButtonBlockElement{
							Type:     "button",
							Text:     slack.NewTextBlockObject("plain_text", ":link:", true, false),
							ActionID: "button-action",
							URL:      link,
							Value:    "More",
						},
					},
				),
			},
		},
	}
}

// TitleEnrichWithEmojiByNotificationType return a title with emoji
// base on notification type
func TitleEnrichWithEmojiByNotificationType(notificationType string) string {
	switch notificationType {
	case "new_post":
		return ":new:"
	case "new_comment":
		return ":speech_balloon:"
	case "change_status":
		return ":up:"
	case "delete_post":
		return ":x:"
	}

	return ""
}
