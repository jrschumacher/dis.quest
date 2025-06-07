// Package logger provides logging utilities with structured logging support
package logger

// Info logs an info message using the default logger.
func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

// Error logs an error message using the default logger.
func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}
