package custom_test

import (
	"context"
	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/enum"
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/services/httpclient/httpclientmock"
	"github.com/getfider/fider/app/services/webhook/custom"
	"io/ioutil"
	"testing"
)

var ctx context.Context

func reset() {
	ctx = context.WithValue(context.Background(), app.TenantCtxKey, &models.Tenant{
		Subdomain: "got",
	})
	bus.Init(custom.Service{}, httpclientmock.Service{})
}

func TestClient_NotifyCustomSuccess(t *testing.T) {
	RegisterT(t)
	env.Config.Webhook.Custom.URL = "http://localhost:8080/fider"
	reset()

	expected := `{"title":"Title","description":"Description","link":"http://domain.com/link","user":{"id":1,"name":"User","role":"collaborator","status":"active"}}`
	bus.Publish(ctx, &cmd.SendWebhookNotification{
		Type:    "new_post",
		Title:   "Title",
		Link:    "http://domain.com/link",
		Content: "Description",
		User: models.User{
			ID:     1,
			Name:   "User",
			Email:  "user@test.local",
			Role:   enum.RoleCollaborator,
			Status: enum.UserActive,
		},
	})

	Expect(httpclientmock.RequestsHistory).HasLen(1)
	Expect(httpclientmock.RequestsHistory[0].URL.String()).Equals("http://localhost:8080/fider")
	Expect(httpclientmock.RequestsHistory[0].Header.Get("Content-Type")).Equals("application/json")
	bytes, err := ioutil.ReadAll(httpclientmock.RequestsHistory[0].Body)
	Expect(err).IsNil()
	Expect(string(bytes)).Equals(expected)
}
