package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/log"
	"strings"
)

// Notifier send a notification to a webhook
type Notifier interface {
	Notify(ctx context.Context, message Message) error
}

// Message represent a JSON message
type Message json.RawMessage

// Client represent configuration of Notifier
type Client struct {
	Notifier string
	URL      string
}

// Notify send a notification message to a webhook url
func (c *Client) Notify(ctx context.Context, message Message) error {
	req := &cmd.HTTPRequest{
		Method: "POST",
		URL:    c.URL,
		Body:   strings.NewReader(string(message)),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	err := bus.Dispatch(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to send notification through %s", c.Notifier)
	}
	log.Debugf(ctx, fmt.Sprintf("Notification sent through %s with response code @{StatusCode} @{Message}.", c.Notifier), dto.Props{
		"StatusCode": req.ResponseStatusCode,
		"Message":    string(req.ResponseBody),
	})

	return nil
}
