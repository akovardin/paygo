package handlers

import (
	"github.com/labstack/echo/v5"
)

type Home struct {
}

func NewHome() *Home {
	return &Home{}
}

func (h *Home) Home(c echo.Context) error {
	return nil
}
