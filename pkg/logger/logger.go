package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logger  *log.Logger
	file    *os.File
	once    sync.Once
	logPath string
)

// Init initializes the internal hidden logger.
func Init() {
	once.Do(func() {
		// Determine log directory relative to the executable, not CWD
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "winitrix: warning: could not determine executable path for logging: %v\n", err)
			return
		}

		logDir := filepath.Join(filepath.Dir(exePath), "logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "winitrix: warning: could not create log directory %s: %v\n", logDir, err)
			return
		}

		logFile := filepath.Join(logDir, "winitrix.log")
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "winitrix: warning: could not open log file %s: %v\n", logFile, err)
			return
		}
		file = f
		logPath = logFile
		logger = log.New(file, "", 0)
		logger.Printf("=== Winitrix Session Started at %s ===\n", time.Now().Format(time.RFC3339))
	})
}

// Info logs an informational message.
func Info(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] %s: %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, v...))
	}
}

// Error logs an error message.
func Error(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] %s: %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, v...))
	}
}

// Debug logs a debug message with raw output.
func Debug(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[DEBUG] %s: %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, v...))
	}
}

// Close closes the log file.
func Close() {
	if file != nil {
		logger.Printf("=== Winitrix Session Ended at %s ===\n", time.Now().Format(time.RFC3339))
		file.Close()
	}
}

// LogPath returns the log file path when available.
func LogPath() string {
	return logPath
}
