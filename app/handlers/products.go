package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

type Products struct {
	app *pocketbase.PocketBase
}

func NewProducts(app *pocketbase.PocketBase) *Products {
	return &Products{
		app: app,
	}
}

type ListResponse struct {
	Products []Product `json:"products"`
}

type Product struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func (h *Products) List(c echo.Context) error {
	id := c.PathParam("app")

	app, err := h.app.Dao().FindRecordById("applications", id)
	if err != nil {
		return err
	}

	products, err := h.app.Dao().FindRecordsByFilter(
		"products",
		"enabled = true && application = {:app}",
		"-created",
		100,
		0,
		dbx.Params{
			"app": app.Id,
		},
	)

	if err != nil {
		return err
	}

	resp := ListResponse{
		Products: []Product{},
	}

	for _, p := range products {
		resp.Products = append(resp.Products, Product{
			Id:          p.Id,
			Name:        p.GetString("name"),
			Description: p.GetString("description"),
			Price:       p.GetFloat("price"),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
