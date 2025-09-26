// Package logger provides structured logging capabilities for the image generation application.
// It wraps Go's standard slog package to provide consistent logging across the application
// with support for different log levels and output formats.
package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

var (
	// Default logger instance used throughout the application
	defaultLogger *slog.Logger
)

func init() {
	// Initialize with a sensible default
	defaultLogger = NewLogger(InfoLevel, false)
}

// LogLevel represents the severity of a log message
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
)

// NewLogger creates a new structured logger with the specified configuration
func NewLogger(level LogLevel, jsonFormat bool) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: toSlogLevel(level),
		AddSource: level == DebugLevel,
	}

	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// SetDefault sets the default logger for the application
func SetDefault(logger *slog.Logger) {
	defaultLogger = logger
	slog.SetDefault(logger)
}

// SetLevel updates the log level of the default logger
func SetLevel(level LogLevel) {
	defaultLogger = NewLogger(level, false)
	slog.SetDefault(defaultLogger)
}

// toSlogLevel converts our LogLevel to slog.Level
func toSlogLevel(level LogLevel) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseLevel parses a string log level
func ParseLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// WithContext returns a logger with context values
func WithContext(ctx context.Context) *slog.Logger {
	return defaultLogger.With("trace_id", ctx.Value("trace_id"))
}

// WithFields returns a logger with additional fields
func WithFields(fields map[string]interface{}) *slog.Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return defaultLogger.With(args...)
}

// WithError returns a logger with an error field
func WithError(err error) *slog.Logger {
	return defaultLogger.With("error", err.Error())
}

// Helper functions for the default logger

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	defaultLogger.Debug(sprintf(format, args...), "source", sprintf("%s:%d", file, line))
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	defaultLogger.Info(sprintf(format, args...))
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warn(sprintf(format, args...))
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	defaultLogger.Error(sprintf(format, args...))
}

// Fatal logs an error message and exits the program
func Fatal(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// Fatalf logs a formatted error message and exits the program
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Error(sprintf(format, args...))
	os.Exit(1)
}

// sprintf is a helper function for formatting
func sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return strings.TrimSpace(strings.ReplaceAll(format, "\n", " "))
}