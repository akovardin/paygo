package handlers

import "github.com/labstack/echo/v5"

type Payments struct {
}

func NewPayments() *Payments {
	return &Payments{}
}

// List выводит список покупок
func (p *Payments) List(c echo.Context) error {
	// token
	return nil
}

// Payment выводит форму для покупки товара
func (p *Payments) Payment(c echo.Context) error {
	// token
	return nil
}

// Buy обрабатываем нажатие кнопки "купить"
func (p *Payments) Buy(c echo.Context) error {
	// token
	return nil
}
