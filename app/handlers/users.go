package handlers

import "github.com/labstack/echo/v5"

type Users struct {
}

func NewUsers() *Users {
	return &Users{}
}

// Form выводим форму для указания почты
func (u *Users) Login(c echo.Context) error {
	return nil
}

// Send отправляет OTP на почту
func (u *Users) Send(c echo.Context) error {
	return nil
}

// Code выводит форму для проверки кода
func (u *Users) Code(c echo.Context) error {
	return nil
}

// Check выполняет проверку кода и отдает токен
func (u *Users) Check(c echo.Context) error {
	return nil
}
