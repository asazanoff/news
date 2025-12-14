package middleware

import (
	"strings"

	"service/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(validToken string, log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			log.WarnCtx(c, "Missing authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.WarnCtx(c, "Invalid authorization header format")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		token := parts[1]

		if token != validToken {
			log.WarnCtx(c, "Invalid authorization token")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization token",
			})
		}

		return c.Next()
	}
}
