package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/log"
	"github.com/getfider/fider/app/pkg/web"

	_ "github.com/getfider/fider/app/services/billing"
	_ "github.com/getfider/fider/app/services/blob/fs"
	_ "github.com/getfider/fider/app/services/blob/s3"
	_ "github.com/getfider/fider/app/services/blob/sql"
	_ "github.com/getfider/fider/app/services/email/mailgun"
	_ "github.com/getfider/fider/app/services/email/smtp"
	_ "github.com/getfider/fider/app/services/httpclient"
	_ "github.com/getfider/fider/app/services/log/console"
	_ "github.com/getfider/fider/app/services/log/sql"
	_ "github.com/getfider/fider/app/services/oauth"
	_ "github.com/getfider/fider/app/services/sqlstore/postgres"
	_ "github.com/getfider/fider/app/services/webhook/custom"
	_ "github.com/getfider/fider/app/services/webhook/slack"
)

//RunServer starts the Fider Server
//Returns an exitcode, 0 for OK and 1 for ERROR
func RunServer(settings *models.SystemSettings) int {
	svcs := bus.Init()
	ctx := log.WithProperty(context.Background(), log.PropertyKeyTag, "BOOTSTRAP")
	for _, s := range svcs {
		log.Debugf(ctx, "Service '@{ServiceCategory}.@{ServiceName}' has been initialized.", dto.Props{
			"ServiceCategory": s.Category(),
			"ServiceName":     s.Name(),
		})
	}

	bus.Publish(ctx, &cmd.PurgeExpiredNotifications{})

	e := routes(web.New(settings))

	go e.Start(":" + env.Config.Port)
	return listenSignals(e, settings)
}

func listenSignals(e *web.Engine, settings *models.SystemSettings) int {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, append([]os.Signal{syscall.SIGTERM, syscall.SIGINT}, extraSignals...)...)
	for {
		s := <-signals
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			err := e.Stop()
			if err != nil {
				return 1
			}
			return 0
		default:
			ret := handleExtraSignal(s, e, settings)
			if ret >= 0 {
				return ret
			}
		}
	}
}
