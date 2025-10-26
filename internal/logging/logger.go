package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var globalLogger *logrus.Logger

// Config represents logging configuration
type Config struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // text, json
	Output string `yaml:"output"` // stdout, file path, or "stdout,/path/to/file.log"
}

// InitLogger initializes the global logger based on configuration
func InitLogger(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}
	}

	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}
	logger.SetLevel(level)

	// Set log format
	switch strings.ToLower(cfg.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text", "":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		return fmt.Errorf("invalid log format %q (must be 'text' or 'json')", cfg.Format)
	}

	// Set log output(s)
	outputs := strings.Split(cfg.Output, ",")
	var writers []io.Writer

	for _, output := range outputs {
		output = strings.TrimSpace(output)
		if output == "" {
			continue
		}

		if output == "stdout" {
			writers = append(writers, os.Stdout)
		} else if output == "stderr" {
			writers = append(writers, os.Stderr)
		} else {
			// It's a file path - create directory if needed
			dir := filepath.Dir(output)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory %q: %w", dir, err)
			}

			// Open file in append mode
			file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("failed to open log file %q: %w", output, err)
			}

			writers = append(writers, file)
		}
	}

	// If no valid outputs, default to stdout
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	// Set output to multi-writer if multiple outputs
	if len(writers) == 1 {
		logger.SetOutput(writers[0])
	} else {
		logger.SetOutput(io.MultiWriter(writers...))
	}

	globalLogger = logger
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if globalLogger == nil {
		// Initialize with defaults if not already initialized
		_ = InitLogger(nil)
	}
	return globalLogger
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// WithField returns a logger with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields returns a logger with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}
