package handlers

import (
	handler "service/internal/handlers/news"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, newsHandler handler.NewsHandler, middlewares ...fiber.Handler) {
	api := app.Group("/", middlewares...)

	api.Post("edit/:id", newsHandler.EditNews)
	api.Get("list", newsHandler.ListNews)
	api.Post("create", newsHandler.CreateNews)
}
