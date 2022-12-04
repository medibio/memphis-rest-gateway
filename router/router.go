package router

import (
	"http-proxy/middlewares"
	"http-proxy/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/memphisdev/memphis.go"
)

// SetupRoutes setup router api
func SetupRoutes(conn *memphis.Conn) *fiber.App {
	utils.InitializeValidations()
	app := fiber.New()
	app.Use(cors.New())
	app.Use(middlewares.Authenticate)
	InitilizeAuthRoutes(app)
	InitializeStationsRoutes(app, conn)

	return app
}
