package config

import (
	"time"

	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestLog struct {
	RequestID string `json:"request_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Duration  string `json:"duration"`

	UserID int    `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
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

	// Default untuk request yang belum terautentikasi.
	userID := 0
	role := ""

	// RequireAuth menyimpan AuthUser ke Locals("user").
	if user, ok := c.Locals("user").(model.AuthUser); ok {
		userID = user.UserID
		role = user.Role
	}

	data := RequestLog{
		RequestID: requestID,
		Method:    c.Method(),
		Path:      c.Path(),
		Status:    c.Response().StatusCode(),
		Duration:  time.Since(start).String(),
		UserID:    userID,
		Role:      role,
	}

	if m.logFunc != nil {
		m.logFunc(data)
	}

	return err
}