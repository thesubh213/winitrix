package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// mu guards logger, file, and logPath: worker goroutines may keep logging
	// while the command that started them is shutting down.
	mu        sync.RWMutex
	logger    *log.Logger
	file      *os.File
	once      sync.Once
	closeOnce sync.Once
	logPath   string
	debugOn   atomic.Bool
)

// SetDebug enables or disables debug-level logging.
func SetDebug(enabled bool) {
	debugOn.Store(enabled)
}

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

		mu.Lock()
		file = f
		logPath = logFile
		logger = log.New(f, "", 0)
		mu.Unlock()

		Info("=== Winitrix Session Started at %s ===", time.Now().Format(time.RFC3339))
	})
}

func logf(level, format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if logger == nil {
		return
	}
	logger.Printf("[%s] %s: %s\n", level, time.Now().Format(time.RFC3339), fmt.Sprintf(format, v...))
}

// Info logs an informational message.
func Info(format string, v ...interface{}) {
	logf("INFO", format, v...)
}

// Error logs an error message.
func Error(format string, v ...interface{}) {
	logf("ERROR", format, v...)
}

// Debug logs a debug message with raw output. Debug logging is opt-in.
func Debug(format string, v ...interface{}) {
	if !debugOn.Load() {
		return
	}
	logf("DEBUG", format, v...)
}

// Close closes the log file. It is safe to call multiple times and safe to
// call while other goroutines are still logging.
func Close() {
	closeOnce.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		if logger != nil && file != nil {
			logger.Printf("=== Winitrix Session Ended at %s ===\n", time.Now().Format(time.RFC3339))
		}
		if file != nil {
			_ = file.Close()
		}
		logger = nil
		file = nil
	})
}

// LogPath returns the log file path when available.
func LogPath() string {
	mu.RLock()
	defer mu.RUnlock()
	return logPath
}
