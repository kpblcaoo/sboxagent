package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/kpblcaoo/sboxagent/internal/utils"
)

// MonitorService represents the status monitoring service
type MonitorService struct {
	config *config.Config
	logger *logger.Logger

	// State
	mu        sync.RWMutex
	running   bool
	startTime time.Time

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Monitoring data
	metrics map[string]interface{}
	alerts  []Alert
}

// Alert represents a monitoring alert
type Alert struct {
	ID        string                 `json:"id"`
	Level     string                 `json:"level"` // info, warning, error, critical
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// MonitorConfig represents monitoring configuration
type MonitorConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Interval       string `mapstructure:"interval"`
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	AlertsEnabled  bool   `mapstructure:"alerts_enabled"`
	RetentionDays  int    `mapstructure:"retention_days"`
}

// NewMonitorService creates a new monitoring service
func NewMonitorService(cfg *config.Config, log *logger.Logger) (*MonitorService, error) {
	return &MonitorService{
		config:  cfg,
		logger:  log,
		metrics: make(map[string]interface{}),
		alerts:  make([]Alert, 0),
	}, nil
}

// Start starts the monitoring service
func (m *MonitorService) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("monitoring service is already running")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true
	m.startTime = time.Now()

	m.logger.Info("Monitoring service starting", map[string]interface{}{
		"metrics_enabled": m.config.Services.Monitoring.MetricsEnabled,
		"alerts_enabled":  m.config.Services.Monitoring.AlertsEnabled,
	})

	// Start monitoring loop
	m.wg.Add(1)
	go m.monitoringLoop()

	return nil
}

// Stop stops the monitoring service
func (m *MonitorService) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.logger.Info("Monitoring service stopping", map[string]interface{}{})
	m.cancel()
	m.wg.Wait()
	m.running = false
}

// monitoringLoop is the main monitoring loop
func (m *MonitorService) monitoringLoop() {
	defer m.wg.Done()

	// Parse interval
	interval, err := utils.ParseDuration(m.config.Services.Monitoring.Interval)
	if err != nil {
		m.logger.Error("Invalid monitoring interval format", map[string]interface{}{
			"interval": m.config.Services.Monitoring.Interval,
			"error":    err.Error(),
		})
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial monitoring
	m.collectMetrics()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("Monitoring loop stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// collectMetrics collects system and service metrics
func (m *MonitorService) collectMetrics() {
	metrics := make(map[string]interface{})

	// System metrics
	metrics["system"] = m.collectSystemMetrics()

	// Service metrics
	metrics["services"] = m.collectServiceMetrics()

	// Performance metrics
	metrics["performance"] = m.collectPerformanceMetrics()

	// Update metrics
	m.mu.Lock()
	m.metrics = metrics
	m.mu.Unlock()

	m.logger.Debug("Metrics collected", map[string]interface{}{
		"metrics_count": len(metrics),
	})
}

// collectSystemMetrics collects system-level metrics
func (m *MonitorService) collectSystemMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Uptime
	metrics["uptime"] = time.Since(m.startTime).String()

	// Memory usage (simplified)
	metrics["memory_usage"] = "unknown" // Would use runtime.ReadMemStats in real implementation

	// CPU usage (simplified)
	metrics["cpu_usage"] = "unknown" // Would use system calls in real implementation

	// Disk usage (simplified)
	metrics["disk_usage"] = "unknown" // Would use system calls in real implementation

	return metrics
}

// collectServiceMetrics collects service-level metrics
func (m *MonitorService) collectServiceMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Service status
	metrics["services_running"] = 0
	metrics["services_total"] = 0

	// This would be populated by actual service status checks
	// For now, we'll use placeholder data

	return metrics
}

// collectPerformanceMetrics collects performance metrics
func (m *MonitorService) collectPerformanceMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Response times
	metrics["avg_response_time"] = "unknown"

	// Throughput
	metrics["requests_per_second"] = "unknown"

	// Error rates
	metrics["error_rate"] = "unknown"

	return metrics
}

// AddAlert adds a new alert
func (m *MonitorService) AddAlert(level, message string, data map[string]interface{}) {
	alert := Alert{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}

	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	m.mu.Unlock()

	m.logger.Info("Alert added", map[string]interface{}{
		"level":   level,
		"message": message,
		"data":    data,
	})
}

// GetMetrics returns the current metrics
func (m *MonitorService) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]interface{})
	for k, v := range m.metrics {
		metrics[k] = v
	}

	return metrics
}

// GetAlerts returns the current alerts
func (m *MonitorService) GetAlerts() []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]Alert, len(m.alerts))
	copy(alerts, m.alerts)

	return alerts
}

// ClearAlerts clears all alerts
func (m *MonitorService) ClearAlerts() {
	m.mu.Lock()
	m.alerts = make([]Alert, 0)
	m.mu.Unlock()

	m.logger.Info("Alerts cleared", map[string]interface{}{})
}

// GetStatus returns the current service status
func (m *MonitorService) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := map[string]interface{}{
		"running":         m.running,
		"startTime":       m.startTime,
		"uptime":          time.Since(m.startTime).String(),
		"metrics_enabled": m.config.Services.Monitoring.MetricsEnabled,
		"alerts_enabled":  m.config.Services.Monitoring.AlertsEnabled,
		"alerts_count":    len(m.alerts),
		"metrics_count":   len(m.metrics),
	}

	return status
}

// IsRunning returns true if the service is running
func (m *MonitorService) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetHealthStatus returns the overall health status
func (m *MonitorService) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(m.startTime).String(),
	}

	// Check for critical alerts
	alerts := m.GetAlerts()
	criticalAlerts := 0
	for _, alert := range alerts {
		if alert.Level == "critical" {
			criticalAlerts++
		}
	}

	if criticalAlerts > 0 {
		status["status"] = "unhealthy"
		status["critical_alerts"] = criticalAlerts
	}

	return status
}
