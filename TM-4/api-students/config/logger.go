package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger     *log.Logger
	logFile    *os.File
	loggerOnce sync.Once
)

func InitLogger() error {

	var err error

	loggerOnce.Do(func() {

		if err = os.MkdirAll("logs", 0755); err != nil {
			return
		}

		logPath := filepath.Join("logs", "app.log")

		logFile, err = os.OpenFile(
			logPath,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0644,
		)

		if err != nil {
			return
		}

		writer := io.MultiWriter(
			os.Stdout,
			logFile,
		)

		logger = log.New(
			writer,
			"",
			0,
		)
	})

	return err
}

func GetLogger() (*log.Logger, *os.File) {
	return logger, logFile
}

func CloseLogger() {

	if logFile != nil {
		_ = logFile.Close()
	}
}
