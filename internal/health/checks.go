package health

import (
	"context"
	"runtime"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/kpblcaoo/sboxagent/internal/services"
)

// DispatcherStats interface for dispatcher statistics
type DispatcherStats interface {
	GetEventsProcessed() int64
	GetEventsDropped() int64
	GetErrors() int64
	GetLastEventTime() time.Time
}

// AggregatorStats interface for aggregator statistics
type AggregatorStats interface {
	GetTotalEntries() int64
	GetDroppedEntries() int64
	GetCurrentEntries() int64
	GetNewestEntry() time.Time
}

// SystemHealthCheck checks system resources
type SystemHealthCheck struct {
	logger *logger.Logger
	name   string
}

// NewSystemHealthCheck creates a new system health check
func NewSystemHealthCheck(log *logger.Logger) *SystemHealthCheck {
	return &SystemHealthCheck{
		logger: log,
		name:   "system",
	}
}

// Name returns the check name
func (h *SystemHealthCheck) Name() string {
	return h.name
}

// Check performs the system health check
func (h *SystemHealthCheck) Check(ctx context.Context) ComponentHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory usage percentage (rough estimate)
	memUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100

	// Determine status based on thresholds
	var status HealthStatus
	var message string

	switch {
	case memUsagePercent > 90:
		status = HealthStatusUnhealthy
		message = "Memory usage is critically high"
	case memUsagePercent > 75:
		status = HealthStatusDegraded
		message = "Memory usage is high"
	default:
		status = HealthStatusHealthy
		message = "System resources are healthy"
	}

	return ComponentHealth{
		Name:      h.name,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"memory_usage_percent": memUsagePercent,
			"memory_alloc":         m.Alloc,
			"memory_sys":           m.Sys,
			"goroutines":           runtime.NumGoroutine(),
			"cpu_count":            runtime.NumCPU(),
		},
	}
}

// SboxctlHealthCheck checks sboxctl service health
type SboxctlHealthCheck struct {
	logger  *logger.Logger
	name    string
	service *services.SboxctlService
}

// NewSboxctlHealthCheck creates a new sboxctl health check
func NewSboxctlHealthCheck(log *logger.Logger, service *services.SboxctlService) *SboxctlHealthCheck {
	return &SboxctlHealthCheck{
		logger:  log,
		name:    "sboxctl",
		service: service,
	}
}

// Name returns the check name
func (h *SboxctlHealthCheck) Name() string {
	return h.name
}

// Check performs the sboxctl health check
func (h *SboxctlHealthCheck) Check(ctx context.Context) ComponentHealth {
	if h.service == nil {
		return ComponentHealth{
			Name:      h.name,
			Status:    HealthStatusUnknown,
			Message:   "Sboxctl service not available",
			Timestamp: time.Now(),
		}
	}

	status := h.service.GetStatus()

	// Extract status information
	running, _ := status["running"].(bool)
	lastRun, _ := status["lastRun"].(time.Time)
	lastError, hasError := status["lastError"].(string)

	var healthStatus HealthStatus
	var message string

	if !running {
		healthStatus = HealthStatusUnhealthy
		message = "Sboxctl service is not running"
	} else if hasError && lastError != "" {
		healthStatus = HealthStatusDegraded
		message = "Sboxctl service has errors"
	} else if time.Since(lastRun) > 5*time.Minute {
		healthStatus = HealthStatusDegraded
		message = "Sboxctl service hasn't run recently"
	} else {
		healthStatus = HealthStatusHealthy
		message = "Sboxctl service is healthy"
	}

	return ComponentHealth{
		Name:      h.name,
		Status:    healthStatus,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"running":          running,
			"lastRun":          lastRun,
			"lastError":        lastError,
			"timeSinceLastRun": time.Since(lastRun),
		},
	}
}

// DispatcherHealthCheck checks event dispatcher health
type DispatcherHealthCheck struct {
	logger     *logger.Logger
	name       string
	dispatcher DispatcherStats
}

// NewDispatcherHealthCheck creates a new dispatcher health check
func NewDispatcherHealthCheck(log *logger.Logger, dispatcher DispatcherStats) *DispatcherHealthCheck {
	return &DispatcherHealthCheck{
		logger:     log,
		name:       "dispatcher",
		dispatcher: dispatcher,
	}
}

// Name returns the check name
func (h *DispatcherHealthCheck) Name() string {
	return h.name
}

// Check performs the dispatcher health check
func (h *DispatcherHealthCheck) Check(ctx context.Context) ComponentHealth {
	if h.dispatcher == nil {
		return ComponentHealth{
			Name:      h.name,
			Status:    HealthStatusUnknown,
			Message:   "Dispatcher not available",
			Timestamp: time.Now(),
		}
	}

	// Calculate error rate
	var errorRate float64
	eventsProcessed := h.dispatcher.GetEventsProcessed()
	if eventsProcessed > 0 {
		errorRate = float64(h.dispatcher.GetErrors()) / float64(eventsProcessed) * 100
	}

	// Calculate drop rate
	var dropRate float64
	if eventsProcessed > 0 {
		dropRate = float64(h.dispatcher.GetEventsDropped()) / float64(eventsProcessed) * 100
	}

	// Determine status based on thresholds
	var status HealthStatus
	var message string

	switch {
	case errorRate > 10:
		status = HealthStatusUnhealthy
		message = "High error rate in event processing"
	case errorRate > 5 || dropRate > 5:
		status = HealthStatusDegraded
		message = "Elevated error or drop rate"
	case time.Since(h.dispatcher.GetLastEventTime()) > 10*time.Minute:
		status = HealthStatusDegraded
		message = "No recent events processed"
	default:
		status = HealthStatusHealthy
		message = "Event dispatcher is healthy"
	}

	return ComponentHealth{
		Name:      h.name,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"eventsProcessed":    eventsProcessed,
			"eventsDropped":      h.dispatcher.GetEventsDropped(),
			"errors":             h.dispatcher.GetErrors(),
			"errorRate":          errorRate,
			"dropRate":           dropRate,
			"lastEventTime":      h.dispatcher.GetLastEventTime(),
			"timeSinceLastEvent": time.Since(h.dispatcher.GetLastEventTime()),
		},
	}
}

// AggregatorHealthCheck checks log aggregator health
type AggregatorHealthCheck struct {
	logger     *logger.Logger
	name       string
	aggregator AggregatorStats
}

// NewAggregatorHealthCheck creates a new aggregator health check
func NewAggregatorHealthCheck(log *logger.Logger, aggregator AggregatorStats) *AggregatorHealthCheck {
	return &AggregatorHealthCheck{
		logger:     log,
		name:       "aggregator",
		aggregator: aggregator,
	}
}

// Name returns the check name
func (h *AggregatorHealthCheck) Name() string {
	return h.name
}

// Check performs the aggregator health check
func (h *AggregatorHealthCheck) Check(ctx context.Context) ComponentHealth {
	if h.aggregator == nil {
		return ComponentHealth{
			Name:      h.name,
			Status:    HealthStatusUnknown,
			Message:   "Aggregator not available",
			Timestamp: time.Now(),
		}
	}

	// Calculate drop rate
	var dropRate float64
	totalEntries := h.aggregator.GetTotalEntries()
	if totalEntries > 0 {
		dropRate = float64(h.aggregator.GetDroppedEntries()) / float64(totalEntries) * 100
	}

	// Determine status based on thresholds
	var status HealthStatus
	var message string

	switch {
	case dropRate > 10:
		status = HealthStatusUnhealthy
		message = "High log entry drop rate"
	case dropRate > 5:
		status = HealthStatusDegraded
		message = "Elevated log entry drop rate"
	case time.Since(h.aggregator.GetNewestEntry()) > 5*time.Minute:
		status = HealthStatusDegraded
		message = "No recent log entries"
	default:
		status = HealthStatusHealthy
		message = "Log aggregator is healthy"
	}

	return ComponentHealth{
		Name:      h.name,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"totalEntries":         totalEntries,
			"currentEntries":       h.aggregator.GetCurrentEntries(),
			"droppedEntries":       h.aggregator.GetDroppedEntries(),
			"dropRate":             dropRate,
			"newestEntry":          h.aggregator.GetNewestEntry(),
			"timeSinceNewestEntry": time.Since(h.aggregator.GetNewestEntry()),
		},
	}
}

// ProcessHealthCheck checks the overall process health
type ProcessHealthCheck struct {
	logger    *logger.Logger
	name      string
	startTime time.Time
}

// NewProcessHealthCheck creates a new process health check
func NewProcessHealthCheck(log *logger.Logger, startTime time.Time) *ProcessHealthCheck {
	return &ProcessHealthCheck{
		logger:    log,
		name:      "process",
		startTime: startTime,
	}
}

// Name returns the check name
func (h *ProcessHealthCheck) Name() string {
	return h.name
}

// Check performs the process health check
func (h *ProcessHealthCheck) Check(ctx context.Context) ComponentHealth {
	uptime := time.Since(h.startTime)

	// Determine status based on uptime
	var status HealthStatus
	var message string

	switch {
	case uptime < 30*time.Second:
		status = HealthStatusDegraded
		message = "Process recently started"
	default:
		status = HealthStatusHealthy
		message = "Process is running normally"
	}

	return ComponentHealth{
		Name:      h.name,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"uptime":    uptime,
			"startTime": h.startTime,
			"pid":       runtime.NumGoroutine(), // Placeholder for actual PID
		},
	}
}
