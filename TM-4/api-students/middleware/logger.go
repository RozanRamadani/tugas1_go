package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestLogger struct {
	logger *log.Logger
}

type RequestLog struct {
	RequestID string `json:"request_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Duration  string `json:"duration"`
}

func NewRequestLogger(logger *log.Logger) *RequestLogger {
	return &RequestLogger{
		logger: logger,
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

	payload, marshalErr := json.Marshal(data)

	if marshalErr == nil && m.logger != nil {
		m.logger.Println(string(payload))
	}

	return err
}
