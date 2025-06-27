package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
		hasError bool
	}{
		{"debug", DebugLevel, false},
		{"DEBUG", DebugLevel, false},
		{"info", InfoLevel, false},
		{"INFO", InfoLevel, false},
		{"warn", WarnLevel, false},
		{"WARN", WarnLevel, false},
		{"warning", WarnLevel, false},
		{"error", ErrorLevel, false},
		{"ERROR", ErrorLevel, false},
		{"invalid", InfoLevel, true},
		{"", InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLogLevel(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		level    string
		hasError bool
	}{
		{"debug", false},
		{"info", false},
		{"warn", false},
		{"error", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			logger, err := New(tt.level)
			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestLogger_LogLevels(t *testing.T) {
	logger, err := New("info")
	require.NoError(t, err)

	// Test that debug messages are not logged at info level
	logger.Debug("debug message", map[string]interface{}{"key": "value"})
	// This should not produce output, but we can't easily capture it in tests
	// The important thing is that it doesn't panic

	// Test that info messages are logged
	logger.Info("info message", map[string]interface{}{"key": "value"})
	logger.Warn("warn message", map[string]interface{}{"key": "value"})
	logger.Error("error message", map[string]interface{}{"key": "value"})
}

func TestLogger_SetLevel(t *testing.T) {
	logger, err := New("info")
	require.NoError(t, err)

	// Initially at info level
	assert.Equal(t, InfoLevel, logger.GetLevel())

	// Change to debug level
	logger.SetLevel(DebugLevel)
	assert.Equal(t, DebugLevel, logger.GetLevel())

	// Change to error level
	logger.SetLevel(ErrorLevel)
	assert.Equal(t, ErrorLevel, logger.GetLevel())
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{LogLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger, err := New("debug")
	require.NoError(t, err)

	fields := map[string]interface{}{
		"string_field": "value",
		"int_field":    42,
		"bool_field":   true,
		"float_field":  3.14,
	}

	// Test that logging with fields doesn't panic
	logger.Debug("debug with fields", fields)
	logger.Info("info with fields", fields)
	logger.Warn("warn with fields", fields)
	logger.Error("error with fields", fields)
}

func TestLogger_WithEmptyFields(t *testing.T) {
	logger, err := New("info")
	require.NoError(t, err)

	// Test logging with empty fields map
	logger.Info("message with empty fields", map[string]interface{}{})
	logger.Info("message with nil fields", nil)
}

func TestLogger_WithNilLogger(t *testing.T) {
	// This test ensures that our logger methods handle edge cases
	logger, err := New("info")
	require.NoError(t, err)

	// Test that all methods can be called without panicking
	logger.Debug("debug", nil)
	logger.Info("info", nil)
	logger.Warn("warn", nil)
	logger.Error("error", nil)
} 