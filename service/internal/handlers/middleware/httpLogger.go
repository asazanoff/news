package middleware

import (
	"service/pkg/logger"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HTTPLogger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)

		log.WithFields(map[string]interface{}{
			"method":  c.Method(),
			"path":    c.Path(),
			"status":  c.Response().StatusCode(),
			"latency": latency.String(),
		}).Info("HTTP request")

		return err
	}
}
