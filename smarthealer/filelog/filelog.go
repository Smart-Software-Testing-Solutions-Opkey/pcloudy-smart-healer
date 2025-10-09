package filelog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalLogger *FileLogger
	once         sync.Once
)

// FileLogger writes log messages to a log file in the database directory
type FileLogger struct {
	logFile *os.File
	logger  *log.Logger
	mu      sync.Mutex
}

// Init initializes the global file logger
// dbPath should be the full path to the SQLite database file
func Init(dbPath string) error {
	var initErr error
	once.Do(func() {
		// Get the directory containing the database
		dbDir := filepath.Dir(dbPath)

		// Create the log file in the same directory
		logFilePath := filepath.Join(dbDir, "smarthealer.log")

		// Open log file in append mode
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		globalLogger = &FileLogger{
			logFile: file,
			logger:  log.New(file, "", log.LstdFlags),
		}
	})

	return initErr
}

// Info logs an informational message
func Info(format string, args ...interface{}) {
	logInternal("[INFO] "+format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	logInternal("[ERROR] "+format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	logInternal("[WARN] "+format, args...)
}

// logInternal is the internal logging function
func logInternal(format string, args ...interface{}) {
	if globalLogger == nil {
		// Fallback to stderr if logger not initialized
		log.Printf("[SmartHealer] "+format, args...)
		return
	}

	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

	globalLogger.logger.Printf(format, args...)
}

// Close closes the log file
func Close() error {
	if globalLogger == nil {
		return nil
	}

	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

	if globalLogger.logFile != nil {
		return globalLogger.logFile.Close()
	}

	return nil
}
