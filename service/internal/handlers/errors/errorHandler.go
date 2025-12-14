package errors

import (
	"errors"
	"service/internal/apperrors"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"service/pkg/logger"
)

type ErrorResponse struct {
	Success bool
	Error   string `validate:"omitempty"`
}

func ErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal server error"

		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			code = appErr.StatusCode
			message = appErr.Message

			if code >= 500 {
				log.WithFields(logrus.Fields{
					"method": c.Method(),
					"path":   c.Path(),
					"error":  err.Error(),
				}).Error("Internal server error")
			} else {
				log.WithFields(logrus.Fields{
					"method": c.Method(),
					"path":   c.Path(),
					"error":  message,
				}).Warn("Client error")
			}
		} else {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
				message = fiberErr.Message
			}

			log.WithFields(logrus.Fields{
				"method": c.Method(),
				"path":   c.Path(),
				"error":  err.Error(),
			}).Error("Unexpected error")
		}

		return c.Status(code).JSON(ErrorResponse{
			Success: false,
			Error:   message,
		})
	}
}
