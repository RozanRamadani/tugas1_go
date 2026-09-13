package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestLog struct {
	RequestID string
	Method    string
	Path      string
	Status    int
	Duration  string
}

type RequestLogger struct {
	logFunc func(RequestLog)
}

func NewRequestLogger(
	logFunc func(RequestLog),
) *RequestLogger {
	return &RequestLogger{
		logFunc: logFunc,
	}
}

func (m *RequestLogger) Handler(c *fiber.Ctx) error {

	start := time.Now()

	requestID := c.Get("X-Request-ID")

	if requestID == "" {
		requestID = uuid.NewString()
	}

	c.Set("X-Request-ID", requestID)

	err := c.Next()

	data := RequestLog{
		RequestID: requestID,
		Method:    c.Method(),
		Path:      c.Path(),
		Status:    c.Response().StatusCode(),
		Duration:  time.Since(start).String(),
	}

	if m.logFunc != nil {
		m.logFunc(data)
	}

	return err
}
