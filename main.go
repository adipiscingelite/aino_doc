package main

import (
	"document/routes"
	"document/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := routes.Route()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	customValidator := &utils.CustomValidator{Validator: validator.New()}
	e.Validator = customValidator
	// e.Logger.Fatal(e.Start("192.168.1.5:1234"))
	e.Logger.Fatal(e.Start("192.168.110.84:1234"))
	// e.Logger.Fatal(e.Start(":1234"))

}
