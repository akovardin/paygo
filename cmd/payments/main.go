package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/template"
	"go.uber.org/fx"

	"gohome.4gophers.ru/kovardin/payments/app/handlers"
	"gohome.4gophers.ru/kovardin/payments/app/settings"
	_ "gohome.4gophers.ru/kovardin/payments/migrations"
	"gohome.4gophers.ru/kovardin/payments/static"
)

func main() {
	fx.New(
		handlers.Module,
		// tasks.Module,

		fx.Provide(pocketbase.New),
		fx.Provide(template.NewRegistry),
		fx.Provide(settings.New),

		fx.Invoke(
			migration,
		),
		fx.Invoke(
			routing,
		),
	).Run()
}

func routing(
	app *pocketbase.PocketBase,
	lc fx.Lifecycle,
	settings *settings.Settings,
	home *handlers.Home,
	payments *handlers.Payments,
	users *handlers.Users,
) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/", home.Home)

		v1 := e.Router.Group("/v1")

		// payments
		v1.GET("/:app/:product/payments/purchase", payments.Purchase, apis.RequireRecordAuth("users"))
		v1.GET("/:app/:product/payments/success", payments.Success, apis.RequireRecordAuth("users"))
		v1.POST("/payments/confirm", payments.Confirm)

		// users
		v1.GET("/user/login", users.Login)
		v1.POST("/user/send", users.Send)
		v1.POST("/user/:email/check", users.Check)
		v1.GET("/user/success", users.Success)

		// static
		e.Router.GET("/static/*", func(c echo.Context) error {
			p := c.PathParam("*")

			path, err := url.PathUnescape(p)
			if err != nil {
				return fmt.Errorf("failed to unescape path variable: %w", err)
			}

			err = c.FileFS(path, static.FS)
			if err != nil && errors.Is(err, echo.ErrNotFound) {
				return c.FileFS("index.html", static.FS)
			}

			return err
		})

		return nil

	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go app.Start()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

// go run cmd/depot/main.go migrate collections --dir ./data/
func migration(app *pocketbase.PocketBase) {
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})
}
