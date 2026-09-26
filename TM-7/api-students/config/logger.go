package config

import (
	"encoding/json"
	"log"
)

var logger *log.Logger

// InitLogger menginisialisasi logger aplikasi.
func InitLogger() error {
	logger = log.Default()

	return nil
}

// CloseLogger menutup logger.
//
// Logger saat ini menggunakan stdout/stderr milik log package,
// sehingga tidak ada resource file yang perlu ditutup.
func CloseLogger() {
	logger = nil
}

// LogRequest mencatat informasi request dalam format JSON.
func LogRequest(data interface{}) {
	if logger == nil {
		logger = log.Default()
	}

	payload, err := json.Marshal(data)

	if err != nil {
		logger.Printf("request log error: %v", err)
		return
	}

	logger.Printf("%s", payload)
}
