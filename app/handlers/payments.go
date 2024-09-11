package handlers

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/template"

	"gohome.4gophers.ru/kovardin/payments/views"
)

const (
	StatusCreated = "created"
	StatusPaid    = "paid"
	StatusConfirm = "confirm"
)

type Payments struct {
	app      *pocketbase.PocketBase
	registry *template.Registry
}

func NewPayments(app *pocketbase.PocketBase, registry *template.Registry) *Payments {
	return &Payments{
		app:      app,
		registry: registry,
	}
}

// Purchase выводит форму для покупки товара
func (h *Payments) Purchase(c echo.Context) error {
	app := c.PathParam("app")
	id := c.PathParam("product")

	token := c.Request().Header.Get("Authorization")
	user, err := h.app.Dao().FindAuthRecordByToken(token, h.app.Settings().RecordAuthToken.Secret)

	if err != nil {
		return err
	}

	application, err := h.app.Dao().FindFirstRecordByFilter(
		"applications",
		"id = {:id}",
		dbx.Params{"id": app},
	)

	if err != nil {
		return err
	}

	product, err := h.app.Dao().FindFirstRecordByFilter(
		"products",
		"id = {:id} && application = {:app}",
		dbx.Params{"id": id, "app": app},
	)

	if err != nil {
		return err
	}

	payments, err := h.app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return err
	}

	record := models.NewRecord(payments)

	record.Set("name", product.GetString("name"))
	record.Set("description", product.GetString("description"))
	record.Set("status", StatusCreated)
	record.Set("amount", product.GetFloat("price"))
	record.Set("product", product.Id)
	record.Set("user", user.Id)
	record.Set("app", app)

	if err := h.app.Dao().SaveRecord(record); err != nil {
		return err
	}

	label := Label{
		Payment: record.Id,
		Product: record.GetString("product"),
		App:     product.GetString("application"),
	}

	html, err := h.registry.LoadFS(views.FS,
		"layout.html",
		"payments/purchase.html",
	).Render(map[string]any{
		"base":    h.app.Settings().Meta.AppUrl,
		"title":   record.GetString("name"),
		"payment": record.Id,
		"product": record.GetString("product"),
		"amount":  record.GetFloat("amount"),
		"wallet":  application.GetString("wallet"),
		"label":   label.Format(),
		"status":  record.GetString("status"),
	})

	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, html)
}

// Success обрабатываем редирект из yoomoney
func (h *Payments) Success(c echo.Context) error {
	id := c.QueryParam("payment")

	payment, err := h.app.Dao().FindFirstRecordByFilter(
		"payments",
		"id = {:id}",
		dbx.Params{"id": id},
	)

	if err != nil {
		return err
	}

	if payment.GetString("status") == StatusCreated {
		payment.Set("status", StatusPaid)

		if err := h.app.Dao().SaveRecord(payment); err != nil {
			return err
		}
	}

	html, err := h.registry.LoadFS(views.FS,
		"layout.html",
		"payments/success.html",
	).Render(map[string]any{})

	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, html)
}

// Confirm обрабатываем колбек от yoomoney
// Пример колбека notification_type=p2p-incoming&bill_id=&amount=130.43&datetime=2024-09-08T11%3A58%3A25Z&codepro=false&sender=41001000040&sha1_hash=8c8e011f7cf624e0923f959480b04fc4531e6778&test_notification=true&operation_label=&operation_id=test-notification&currency=643&label=
func (h *Payments) Confirm(c echo.Context) error {
	data := struct {
		NotificationType string  `json:"notification_type" form:"notification_type"`
		BillId           string  `json:"bill_id" form:"bill_id"`
		Amount           float64 `json:"amount" form:"amount"`
		DateTime         string  `json:"datetime" form:"datetime"`
		Codepro          bool    `json:"codepro" form:"codepro"`
		Sender           string  `json:"sender" form:"sender"`
		Sha1Hash         string  `json:"sha1_hash" form:"sha1_hash"`
		OperationLabel   string  `json:"operation_label" form:"operation_label"`
		OperationId      string  `json:"operation_id" form:"operation_id"`
		Currency         int     `json:"currency" form:"currency"`
		Label            string  `json:"label" form:"label"`
	}{}
	if err := c.Bind(&data); err != nil {
		h.app.Logger().Error("error on bind login request", "err", err)

		return err
	}

	h.app.Logger().Info("confirm data", "data", data)

	label, err := Label{}.Parse(data.Label)
	if err != nil {
		return err
	}

	app, err := h.app.Dao().FindRecordById("applications", label.App)
	if err != nil {
		h.app.Logger().Error("error on find application", "err", err, "id", label.App)

		return err
	}

	token := app.GetString("secret")
	// notification_type&operation_id&amount&currency&datetime&sender&codepro&notification_secret&label

	check := data.NotificationType +
		"&" + data.OperationId +
		"&" + fmt.Sprintf("%.2f", data.Amount) +
		"&" + fmt.Sprintf("%d", data.Currency) +
		"&" + data.DateTime +
		"&" + data.Sender +
		"&" + fmt.Sprintf("%t", data.Codepro) +
		"&" + token +
		"&" + data.Label

	hash := sha1.New()
	hash.Write([]byte(check))
	sha := hex.EncodeToString(hash.Sum(nil))

	if sha != data.Sha1Hash {
		h.app.Logger().Error("error on check hash", "counted", sha, "income", data.Sha1Hash)

		return errors.New("error on check hash")
	}

	// TODO: check product status

	record, err := h.app.Dao().FindRecordById("payments", label.Payment)
	if err != nil {
		h.app.Logger().Error("error on find payment", "err", err, "id", label.Payment)

		return err
	}

	record.Set("status", StatusConfirm)

	if err := h.app.Dao().SaveRecord(record); err != nil {
		return err
	}

	// token
	return nil
}

type Label struct {
	Payment string
	Product string
	App     string
}

func (l Label) Parse(in string) (Label, error) {
	parts := strings.Split(in, ":")

	if len(parts) < 3 {
		return l, errors.New("invalid input")
	}

	l.Payment = parts[0]
	l.Product = parts[1]
	l.App = parts[2]

	return l, nil
}

func (l Label) Format() string {
	return l.Payment + ":" + l.Product + ":" + l.App
}
