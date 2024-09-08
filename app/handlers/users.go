package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/pocketbase/pocketbase/tools/template"

	"gohome.4gophers.ru/kovardin/payments/pkg/mail"
	"gohome.4gophers.ru/kovardin/payments/pkg/utils"
	"gohome.4gophers.ru/kovardin/payments/views"
)

type Users struct {
	app      *pocketbase.PocketBase
	registry *template.Registry
}

func NewUsers(app *pocketbase.PocketBase, registry *template.Registry) *Users {

	return &Users{
		app:      app,
		registry: registry,
	}
}

// Form выводим форму для указания почты
func (h *Users) Login(c echo.Context) error {
	product := c.PathParam("product")

	// TODO: check id app and product active

	html, err := h.registry.LoadFS(views.FS,
		"layout.html",
		"users/login.html",
	).Render(map[string]any{
		"product": product,
	})

	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, html)
}

// Send отправляет OTP на почту
func (h *Users) Send(c echo.Context) error {
	product := c.PathParam("product")

	// TODO: check app and product active

	data := struct {
		Email string `json:"email" form:"email"`
	}{}
	if err := c.Bind(&data); err != nil {
		h.app.Logger().Error("error on bind login request", "err", err)

		return err
	}

	record, err := h.app.Dao().FindFirstRecordByData("users", "email", data.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.app.Logger().Error("error search user", "err", err)

		return err
	}

	if record == nil {
		users, err := h.app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			h.app.Logger().Error("error search user collection", "err", err)

			return err
		}

		record = models.NewRecord(users)
		record.Set("email", data.Email)
		record.Set("username", data.Email)
	}

	password := utils.RandomString(6, true, false, true)
	smtp := h.app.Settings().Smtp
	mailer := mail.New(mail.Config{
		Password: smtp.Password,
		Out:      smtp.Host,
		Port:     smtp.Port,
		Username: smtp.Username,
	})

	err = mailer.Send(mail.Message{
		From:    h.app.Settings().Meta.SenderAddress,
		Name:    h.app.Settings().Meta.SenderName,
		To:      data.Email,
		Subject: "Billing password",
		Html:    "<p>Password: <b>" + password + "</b><p>",
	})

	if err != nil {
		h.app.Logger().Error("error on send email", "err", err)

		return err
	}

	record.SetPassword(password)

	if err := h.app.Dao().SaveRecord(record); err != nil {
		h.app.Logger().Error("error on save record", "err", err)

		return err
	}

	html, err := h.registry.LoadFS(views.FS,
		"layout.html",
		"users/code.html",
	).Render(map[string]any{
		"product": product,
		"email":   data.Email,
	})

	if err != nil {
		h.app.Logger().Error("error on render template", "err", err)

		return err
	}

	return c.HTML(http.StatusOK, html)
}

// Check выполняет проверку кода и отдает токен
func (h *Users) Check(c echo.Context) error {
	// app := c.PathParam("app")
	// product := c.PathParam("product")

	email := c.PathParam("email")

	// TODO: check app and product active

	data := struct {
		Password string `json:"password" form:"password"`
	}{}
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("invalid request", err)
	}

	record, err := h.app.Dao().FindFirstRecordByData("users", "email", email)
	if err != nil || !record.ValidatePassword(data.Password) {
		return apis.NewBadRequestError("invalid credentials", err)
	}

	record.SetVerified(true)

	if err := h.app.Dao().SaveRecord(record); err != nil {
		return apis.NewApiError(http.StatusInternalServerError, "not verified", err)
	}

	token, err := tokens.NewRecordAuthToken(h.app, record)
	if err != nil {
		return apis.NewBadRequestError("Failed to create auth token.", err)
	}

	return c.Redirect(http.StatusSeeOther, "/v1/user/success?token="+token)
}

func (h *Users) Success(c echo.Context) error {
	html, err := h.registry.LoadFS(views.FS,
		"layout.html",
		"users/success.html",
	).Render(map[string]any{})

	if err != nil {
		h.app.Logger().Error("error on render template", "err", err)

		return err
	}

	return c.HTML(http.StatusOK, html)
}
