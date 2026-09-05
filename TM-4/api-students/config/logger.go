package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxLogSize int64 = 5 * 1024 * 1024 // 5 MB

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

		if err = openLogFile(); err != nil {
			return
		}
	})

	return err
}

func openLogFile() error {

	logPath := filepath.Join("logs", "app.log")

	var err error

	logFile, err = os.OpenFile(
		logPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return err
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

	return nil
}

// RotateLog melakukan rotasi jika ukuran app.log
// sudah mencapai batas yang ditentukan.
func RotateLog() error {

	if logFile == nil {
		return nil
	}

	info, err := logFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() < maxLogSize {
		return nil
	}

	_ = logFile.Close()

	oldPath := filepath.Join("logs", "app.log")

	newPath := filepath.Join(
		"logs",
		"app-"+time.Now().Format("20060102-150405")+".log",
	)

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	return openLogFile()
}

func GetLogger() (*log.Logger, *os.File) {
	return logger, logFile
}

func CloseLogger() {

	if logFile != nil {
		_ = logFile.Close()
	}
}

func LogRequest(data interface{}) {
	// nanti menangani rotasi + penulisan
}