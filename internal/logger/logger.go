package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// Logger represents a structured logger
type Logger struct {
	level LogLevel
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

// New creates a new logger instance
func New(level string) (*Logger, error) {
	logLevel, err := ParseLogLevel(level)
	if err != nil {
		return nil, err
	}

	return &Logger{
		level: logLevel,
		debug: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warn:  log.New(os.Stderr, "[WARN] ", log.LstdFlags),
		error: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}, nil
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	if l.level <= DebugLevel {
		l.log(l.debug, "DEBUG", message, fields)
	}
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	if l.level <= InfoLevel {
		l.log(l.info, "INFO", message, fields)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	if l.level <= WarnLevel {
		l.log(l.warn, "WARN", message, fields)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	if l.level <= ErrorLevel {
		l.log(l.error, "ERROR", message, fields)
	}
}

// log formats and outputs a log message
func (l *Logger) log(logger *log.Logger, level, message string, fields map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	
	// Build log entry
	entry := fmt.Sprintf("%s [%s] %s", timestamp, level, message)
	
	// Add fields if provided
	if len(fields) > 0 {
		fieldStrs := make([]string, 0, len(fields))
		for key, value := range fields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", key, value))
		}
		entry += " " + strings.Join(fieldStrs, " ")
	}
	
	logger.Println(entry)
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() LogLevel {
	return l.level
} 