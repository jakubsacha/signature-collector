package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Severity levels recognized by Google Cloud Logging
const (
	SeverityDebug    = "DEBUG"
	SeverityInfo     = "INFO"
	SeverityWarning  = "WARNING"
	SeverityError    = "ERROR"
	SeverityCritical = "CRITICAL"
)

// LogEntry represents a structured log entry for Cloud Logging
type LogEntry struct {
	Severity string                 `json:"severity"`
	Message  string                 `json:"message"`
	Time     string                 `json:"time"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides structured logging for Cloud Run
type Logger struct {
	fields map[string]interface{}
}

// New creates a new Logger instance
func New() *Logger {
	return &Logger{
		fields: make(map[string]interface{}),
	}
}

// WithField returns a new logger with the given field added
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		fields: make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// WithFields returns a new logger with the given fields added
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		fields: make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// log writes a structured log entry
func (l *Logger) log(severity, message string) {
	entry := LogEntry{
		Severity: severity,
		Message:  message,
		Time:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	if len(l.fields) > 0 {
		entry.Fields = l.fields
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text if JSON marshaling fails
		fmt.Fprintf(os.Stderr, "%s: %s (json error: %v)\n", severity, message, err)
		return
	}
	fmt.Println(string(jsonBytes))
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(SeverityDebug, message)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(SeverityDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log(SeverityInfo, message)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(SeverityInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(SeverityWarning, message)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(SeverityWarning, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(SeverityError, message)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(SeverityError, fmt.Sprintf(format, args...))
}

// Fatal logs a critical message and exits
func (l *Logger) Fatal(message string) {
	l.log(SeverityCritical, message)
	os.Exit(1)
}

// Fatalf logs a formatted critical message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(SeverityCritical, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Package-level default logger for convenience
var defaultLogger = New()

// Debug logs a debug message using the default logger
func Debug(message string) {
	defaultLogger.Debug(message)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info logs an info message using the default logger
func Info(message string) {
	defaultLogger.Info(message)
}

// Infof logs a formatted info message using the default logger
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(message string) {
	defaultLogger.Warn(message)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error logs an error message using the default logger
func Error(message string) {
	defaultLogger.Error(message)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal logs a critical message using the default logger and exits
func Fatal(message string) {
	defaultLogger.Fatal(message)
}

// Fatalf logs a formatted critical message using the default logger and exits
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// WithField returns a new logger with the given field added
func WithField(key string, value interface{}) *Logger {
	return defaultLogger.WithField(key, value)
}

// WithFields returns a new logger with the given fields added
func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}
