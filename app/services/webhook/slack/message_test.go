package slack_test

import (
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/services/webhook/slack"
	"testing"
)

func TestTitleEnrichWithEmojiByNotificationType_NewPost(t *testing.T) {
	RegisterT(t)

	notificationType := "new_post"
	expected := ":new:"
	title := slack.TitleEnrichWithEmojiByNotificationType(notificationType)

	Expect(title).Equals(expected)
}

func TestTitleEnrichWithEmojiByNotificationType_NewComment(t *testing.T) {
	RegisterT(t)

	notificationType := "new_comment"
	expected := ":speech_balloon:"
	title := slack.TitleEnrichWithEmojiByNotificationType(notificationType)

	Expect(title).Equals(expected)
}

func TestTitleEnrichWithEmojiByNotificationType_ChangeStatus(t *testing.T) {
	RegisterT(t)

	notificationType := "change_status"
	expected := ":up:"
	title := slack.TitleEnrichWithEmojiByNotificationType(notificationType)

	Expect(title).Equals(expected)
}

func TestTitleEnrichWithEmojiByNotificationType_DeletePost(t *testing.T) {
	RegisterT(t)

	notificationType := "delete_post"
	expected := ":x:"
	title := slack.TitleEnrichWithEmojiByNotificationType(notificationType)

	Expect(title).Equals(expected)
}
