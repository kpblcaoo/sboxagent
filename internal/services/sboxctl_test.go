package services

import (
	"context"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSboxctlService(t *testing.T) {
	logger, err := logger.New("info")
	require.NoError(t, err)

	cfg := config.SboxctlConfig{
		Enabled:       true,
		Command:       []string{"echo", "test"},
		Interval:      "1m",
		Timeout:       "30s",
		StdoutCapture: true,
		HealthCheck: config.HealthCheckConfig{
			Enabled:  true,
			Interval: "30s",
			Timeout:  "5s",
		},
	}

	service, err := NewSboxctlService(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"30s", 30 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"1m30s", 90 * time.Second, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			duration, err := parseDuration(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, duration)
			}
		})
	}
}

func TestSboxctlService_ParseEvent(t *testing.T) {
	logger, err := logger.New("info")
	require.NoError(t, err)

	cfg := config.SboxctlConfig{
		Enabled:       true,
		Command:       []string{"echo", "test"},
		Interval:      "1m",
		Timeout:       "30s",
		StdoutCapture: true,
	}

	service, err := NewSboxctlService(cfg, logger)
	require.NoError(t, err)

	// Test valid JSON event
	validJSON := `{"type":"LOG","data":{"level":"info","message":"test"},"timestamp":"2025-06-27T16:30:00Z","version":"1.0"}`
	event, err := service.parseEvent(validJSON)
	require.NoError(t, err)
	assert.Equal(t, "LOG", event.Type)
	assert.Equal(t, "1.0", event.Version)
	assert.Equal(t, "2025-06-27T16:30:00Z", event.Timestamp)
	assert.NotNil(t, event.Data)

	// Test invalid JSON
	invalidJSON := `{"type":"LOG","invalid json`
	_, err = service.parseEvent(invalidJSON)
	assert.Error(t, err)

	// Test missing type
	noTypeJSON := `{"data":{"level":"info"},"timestamp":"2025-06-27T16:30:00Z","version":"1.0"}`
	_, err = service.parseEvent(noTypeJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event type is required")
}

func TestSboxctlService_GetStatus(t *testing.T) {
	logger, err := logger.New("info")
	require.NoError(t, err)

	cfg := config.SboxctlConfig{
		Enabled:       true,
		Command:       []string{"echo", "test"},
		Interval:      "1m",
		Timeout:       "30s",
		StdoutCapture: true,
	}

	service, err := NewSboxctlService(cfg, logger)
	require.NoError(t, err)

	// Get status before starting
	status := service.GetStatus()
	assert.False(t, status["running"].(bool))
	assert.Equal(t, []string{"echo", "test"}, status["command"])
	assert.Equal(t, "1m", status["interval"])
	assert.Equal(t, "30s", status["timeout"])

	// Start service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.Start(ctx)
	require.NoError(t, err)

	// Get status after starting
	status = service.GetStatus()
	assert.True(t, status["running"].(bool))
	assert.NotNil(t, status["lastRun"])

	// Stop service
	service.Stop()
	time.Sleep(100 * time.Millisecond) // Give time for goroutines to stop

	// Get status after stopping
	status = service.GetStatus()
	assert.False(t, status["running"].(bool))
}

func TestSboxctlService_StartStop(t *testing.T) {
	logger, err := logger.New("info")
	require.NoError(t, err)

	cfg := config.SboxctlConfig{
		Enabled:       true,
		Command:       []string{"echo", "test"},
		Interval:      "1m",
		Timeout:       "30s",
		StdoutCapture: true,
	}

	service, err := NewSboxctlService(cfg, logger)
	require.NoError(t, err)

	// Test starting
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.Start(ctx)
	require.NoError(t, err)

	// Test that service is running (check status)
	status := service.GetStatus()
	assert.True(t, status["running"].(bool))

	// Test stopping
	service.Stop()
	time.Sleep(100 * time.Millisecond) // Give time for goroutines to stop

	// Test that service is stopped (check status)
	status = service.GetStatus()
	assert.False(t, status["running"].(bool))
}

func TestSboxctlService_DoubleStart(t *testing.T) {
	logger, err := logger.New("info")
	require.NoError(t, err)

	cfg := config.SboxctlConfig{
		Enabled:       true,
		Command:       []string{"echo", "test"},
		Interval:      "1m",
		Timeout:       "30s",
		StdoutCapture: true,
	}

	service, err := NewSboxctlService(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start first time
	err = service.Start(ctx)
	require.NoError(t, err)

	// Try to start again
	err = service.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestSboxctlService_GetEventChannel(t *testing.T) {
	logger, err := logger.New("info")
	require.NoError(t, err)

	cfg := config.SboxctlConfig{
		Enabled:       true,
		Command:       []string{"echo", "test"},
		Interval:      "1m",
		Timeout:       "30s",
		StdoutCapture: true,
	}

	service, err := NewSboxctlService(cfg, logger)
	require.NoError(t, err)

	// Get event channel
	eventChan := service.GetEventChannel()
	assert.NotNil(t, eventChan)

	// Channel should be buffered - we can't test sending to receive-only channel
	// but we can verify it's not nil and has the right type
	assert.IsType(t, (<-chan SboxctlEvent)(nil), eventChan)
} 