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
	products *handlers.Products,
) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/", home.Home)

		// listing
		e.Router.GET("/:app/products", products.List)

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

	app.OnRecordBeforeUpdateRequest("artifacts").Add(func(e *core.RecordUpdateEvent) error {

		// overwrite the submitted "active" field value to false
		e.Record.Set("active", false)

		// delete all artifact folder

		return nil
	})

	app.OnRecordBeforeUpdateRequest("versions").Add(func(e *core.RecordUpdateEvent) error {

		// overwrite the submitted "active" field value to false
		e.Record.Set("active", false)

		// delete version and update manifest

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
