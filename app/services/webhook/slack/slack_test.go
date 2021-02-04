package slack_test

import (
	"context"
	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/models/cmd"
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/services/httpclient/httpclientmock"
	"github.com/getfider/fider/app/services/webhook/slack"
	"io/ioutil"
	"testing"
)

var ctx context.Context

func reset() {
	ctx = context.WithValue(context.Background(), app.TenantCtxKey, &models.Tenant{
		Subdomain: "got",
	})
	bus.Init(slack.Service{}, httpclientmock.Service{})
}

func TestClient_NotifySlackSuccess(t *testing.T) {
	RegisterT(t)
	env.Config.Webhook.Slack.URL = "https://hooks.slack.com/services/A12345BCD/A01BCDEFGHI/AbcDefg1HIJ2kLmn34oPqRst"
	reset()

	expected := `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":":new: *Title*"}},{"type":"section","text":{"type":"mrkdwn","text":"Description"},"accessory":{"type":"button","text":{"type":"plain_text","text":":link:","emoji":true},"action_id":"button-action","url":"http://domain.com/link","value":"More"}}]}`
	bus.Publish(ctx, &cmd.SendWebhookNotification{
		Type:    "new_post",
		Title:   "Title",
		Link:    "http://domain.com/link",
		Content: "Description",
	})

	Expect(httpclientmock.RequestsHistory).HasLen(1)
	Expect(httpclientmock.RequestsHistory[0].URL.String()).Equals("https://hooks.slack.com/services/A12345BCD/A01BCDEFGHI/AbcDefg1HIJ2kLmn34oPqRst")
	Expect(httpclientmock.RequestsHistory[0].Header.Get("Content-Type")).Equals("application/json")
	bytes, err := ioutil.ReadAll(httpclientmock.RequestsHistory[0].Body)
	Expect(err).IsNil()
	Expect(string(bytes)).Equals(expected)
}
