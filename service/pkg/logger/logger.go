package logger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Logger обертка над logrus для структурированного логирования
type Logger struct {
	*logrus.Logger
}

// New создает новый логгер
func Init() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
	return &Logger{log}
}

// withContext добавляет контекст из fiber.Ctx
func (l *Logger) withContext(c *fiber.Ctx) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"method": c.Method(),
		"path":   c.Path(),
	})
}

// InfoCtx логирует INFO с контекстом запроса
func (l *Logger) InfoCtx(c *fiber.Ctx, message string) {
	l.withContext(c).Info(message)
}

// WarnCtx логирует WARN с контекстом запроса
func (l *Logger) WarnCtx(c *fiber.Ctx, message string) {
	l.withContext(c).Warn(message)
}

// ErrorCtx логирует ERROR с контекстом запроса и ошибкой
func (l *Logger) ErrorCtx(c *fiber.Ctx, message string, err error) {
	l.withContext(c).WithField("error", err.Error()).Error(message)
}

// WarnWithError логирует WARN с ошибкой (для клиентских ошибок)
func (l *Logger) WarnWithError(c *fiber.Ctx, message string, err error) {
	l.withContext(c).WithField("error", err).Warn(message)
}

// ErrorWithFields логирует ERROR с дополнительными полями
func (l *Logger) ErrorWithFields(c *fiber.Ctx, message string, fields logrus.Fields) {
	entry := l.withContext(c)
	for k, v := range fields {
		entry = entry.WithField(k, v)
	}
	entry.Error(message)
}

func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}
