package main

import (
	"document/routes"
	"document/utils"

	// "net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := routes.Route()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	// Serve static files from the "assets" directory
	// e.Static("/assets", "assets")

	// Custom Validator
	customValidator := &utils.CustomValidator{Validator: validator.New()}
	e.Validator = customValidator

	// Start the server
	e.Logger.Fatal(e.Start("192.168.110.9:1234"))
	// e.Logger.Fatal(e.Start("192.168.110.9:1234"))
}
