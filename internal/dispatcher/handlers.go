package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
)

// LogHandler handles log events
type LogHandler struct {
	logger *logger.Logger
	name   string
}

// NewLogHandler creates a new log handler
func NewLogHandler(log *logger.Logger) *LogHandler {
	return &LogHandler{
		logger: log,
		name:   "log_handler",
	}
}

// Handle processes log events
func (h *LogHandler) Handle(ctx context.Context, event Event) error {
	// Extract log level and message
	level, _ := event.Data["level"].(string)
	message, _ := event.Data["message"].(string)

	if message == "" {
		return fmt.Errorf("log event missing message")
	}

	// Log with appropriate level
	switch level {
	case "debug":
		h.logger.Debug(message, event.Data)
	case "info":
		h.logger.Info(message, event.Data)
	case "warn", "warning":
		h.logger.Warn(message, event.Data)
	case "error":
		h.logger.Error(message, event.Data)
	default:
		h.logger.Info(message, event.Data)
	}

	return nil
}

// GetName returns the handler name
func (h *LogHandler) GetName() string {
	return h.name
}

// GetSupportedTypes returns supported event types
func (h *LogHandler) GetSupportedTypes() []EventType {
	return []EventType{EventTypeLog}
}

// ConfigHandler handles configuration events
type ConfigHandler struct {
	logger *logger.Logger
	name   string
	mu     sync.RWMutex
	config map[string]interface{}
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(log *logger.Logger) *ConfigHandler {
	return &ConfigHandler{
		logger: log,
		name:   "config_handler",
		config: make(map[string]interface{}),
	}
}

// Handle processes configuration events
func (h *ConfigHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Update configuration
	for key, value := range event.Data {
		h.config[key] = value
	}

	h.logger.Info("Configuration updated", map[string]interface{}{
		"updatedKeys": len(event.Data),
		"totalKeys":   len(h.config),
	})

	return nil
}

// GetName returns the handler name
func (h *ConfigHandler) GetName() string {
	return h.name
}

// GetSupportedTypes returns supported event types
func (h *ConfigHandler) GetSupportedTypes() []EventType {
	return []EventType{EventTypeConfig}
}

// GetConfig returns the current configuration
func (h *ConfigHandler) GetConfig() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Create a copy to avoid race conditions
	config := make(map[string]interface{})
	for k, v := range h.config {
		config[k] = v
	}
	return config
}

// ErrorHandler handles error events
type ErrorHandler struct {
	logger *logger.Logger
	name   string
	mu     sync.RWMutex
	errors []ErrorRecord
}

// ErrorRecord represents an error record
type ErrorRecord struct {
	Timestamp time.Time              `json:"timestamp"`
	Error     string                 `json:"error"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(log *logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: log,
		name:   "error_handler",
		errors: make([]ErrorRecord, 0),
	}
}

// Handle processes error events
func (h *ErrorHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Extract error message
	errorMsg, _ := event.Data["error"].(string)
	if errorMsg == "" {
		errorMsg = "Unknown error"
	}

	// Create error record
	record := ErrorRecord{
		Timestamp: event.Timestamp,
		Error:     errorMsg,
		Source:    event.Source,
		Data:      event.Data,
	}

	// Add to errors list (keep last 100 errors)
	h.errors = append(h.errors, record)
	if len(h.errors) > 100 {
		h.errors = h.errors[1:]
	}

	h.logger.Error("Error event received", map[string]interface{}{
		"error":       errorMsg,
		"source":      event.Source,
		"totalErrors": len(h.errors),
	})

	return nil
}

// GetName returns the handler name
func (h *ErrorHandler) GetName() string {
	return h.name
}

// GetSupportedTypes returns supported event types
func (h *ErrorHandler) GetSupportedTypes() []EventType {
	return []EventType{EventTypeError}
}

// GetErrors returns the error records
func (h *ErrorHandler) GetErrors() []ErrorRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Create a copy to avoid race conditions
	errors := make([]ErrorRecord, len(h.errors))
	copy(errors, h.errors)
	return errors
}

// StatusHandler handles status events
type StatusHandler struct {
	logger *logger.Logger
	name   string
	mu     sync.RWMutex
	status map[string]interface{}
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(log *logger.Logger) *StatusHandler {
	return &StatusHandler{
		logger: log,
		name:   "status_handler",
		status: make(map[string]interface{}),
	}
}

// Handle processes status events
func (h *StatusHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Update status
	for key, value := range event.Data {
		h.status[key] = value
	}

	h.logger.Info("Status updated", map[string]interface{}{
		"updatedKeys": len(event.Data),
		"totalKeys":   len(h.status),
	})

	return nil
}

// GetName returns the handler name
func (h *StatusHandler) GetName() string {
	return h.name
}

// GetSupportedTypes returns supported event types
func (h *StatusHandler) GetSupportedTypes() []EventType {
	return []EventType{EventTypeStatus}
}

// GetStatus returns the current status
func (h *StatusHandler) GetStatus() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Create a copy to avoid race conditions
	status := make(map[string]interface{})
	for k, v := range h.status {
		status[k] = v
	}
	return status
}

// HealthHandler handles health events
type HealthHandler struct {
	logger *logger.Logger
	name   string
	mu     sync.RWMutex
	health map[string]HealthRecord
}

// HealthRecord represents a health record
type HealthRecord struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: log,
		name:   "health_handler",
		health: make(map[string]HealthRecord),
	}
}

// Handle processes health events
func (h *HealthHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Extract health information
	component, _ := event.Data["component"].(string)
	if component == "" {
		component = "unknown"
	}

	status, _ := event.Data["status"].(string)
	if status == "" {
		status = "unknown"
	}

	// Create health record
	record := HealthRecord{
		Status:    status,
		Timestamp: event.Timestamp,
		Data:      event.Data,
	}

	h.health[component] = record

	h.logger.Info("Health event received", map[string]interface{}{
		"component":       component,
		"status":          status,
		"totalComponents": len(h.health),
	})

	return nil
}

// GetName returns the handler name
func (h *HealthHandler) GetName() string {
	return h.name
}

// GetSupportedTypes returns supported event types
func (h *HealthHandler) GetSupportedTypes() []EventType {
	return []EventType{EventTypeHealth}
}

// GetHealth returns the health records
func (h *HealthHandler) GetHealth() map[string]HealthRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Create a copy to avoid race conditions
	health := make(map[string]HealthRecord)
	for k, v := range h.health {
		health[k] = v
	}
	return health
}
