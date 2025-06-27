package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
)

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth represents the health of a component
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// HealthReport represents a complete health report
type HealthReport struct {
	OverallStatus HealthStatus           `json:"overall_status"`
	Timestamp     time.Time              `json:"timestamp"`
	Components    []ComponentHealth      `json:"components"`
	Summary       map[string]int         `json:"summary"`
	Uptime        time.Duration          `json:"uptime"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// HealthChecker represents the health checker
type HealthChecker struct {
	logger *logger.Logger

	// Configuration
	checkInterval time.Duration
	timeout       time.Duration

	// State
	mu        sync.RWMutex
	running   bool
	startTime time.Time

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Health checks
	checks map[string]HealthCheck

	// Last report
	lastReport HealthReport
	reportMu   sync.RWMutex
}

// HealthCheck defines the interface for health checks
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(log *logger.Logger, checkInterval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		logger:        log,
		checkInterval: checkInterval,
		timeout:       timeout,
		checks:        make(map[string]HealthCheck),
		startTime:     time.Now(),
	}
}

// Start starts the health checker
func (h *HealthChecker) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("health checker is already running")
	}

	h.ctx, h.cancel = context.WithCancel(ctx)
	h.running = true
	h.startTime = time.Now()

	h.logger.Info("Health checker starting", map[string]interface{}{
		"checkInterval": h.checkInterval,
		"timeout":       h.timeout,
		"checks":        len(h.checks),
	})

	// Start health checking loop
	h.wg.Add(1)
	go h.run()

	return nil
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	h.logger.Info("Health checker stopping", map[string]interface{}{})
	h.cancel()
	h.wg.Wait()
	h.running = false
}

// RegisterCheck registers a health check
func (h *HealthChecker) RegisterCheck(check HealthCheck) error {
	if check == nil {
		return fmt.Errorf("health check cannot be nil")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks[check.Name()] = check

	h.logger.Info("Health check registered", map[string]interface{}{
		"name": check.Name(),
	})

	return nil
}

// UnregisterCheck unregisters a health check
func (h *HealthChecker) UnregisterCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.checks, name)

	h.logger.Info("Health check unregistered", map[string]interface{}{
		"name": name,
	})
}

// run is the main health checking loop
func (h *HealthChecker) run() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	// Run initial health check
	h.performHealthCheck()

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info("Health checker loop stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			h.performHealthCheck()
		}
	}
}

// performHealthCheck performs all registered health checks
func (h *HealthChecker) performHealthCheck() {
	h.mu.RLock()
	checks := make(map[string]HealthCheck)
	for k, v := range h.checks {
		checks[k] = v
	}
	h.mu.RUnlock()

	if len(checks) == 0 {
		h.logger.Debug("No health checks registered", map[string]interface{}{})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(h.ctx, h.timeout)
	defer cancel()

	// Limit concurrent checks to prevent DoS
	maxConcurrent := 10
	if len(checks) > maxConcurrent {
		h.logger.Warn("Too many health checks, limiting concurrent execution", map[string]interface{}{
			"total":         len(checks),
			"maxConcurrent": maxConcurrent,
		})
	}

	// Run checks with concurrency limit
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	results := make(chan ComponentHealth, len(checks))

	for name, check := range checks {
		wg.Add(1)
		go func(c HealthCheck, n string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			// Run check with individual timeout
			checkCtx, checkCancel := context.WithTimeout(ctx, h.timeout/2)
			defer checkCancel()

			result := c.Check(checkCtx)
			select {
			case results <- result:
			case <-ctx.Done():
			}
		}(check, name)
	}

	// Wait for all checks to complete with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// All checks completed
	case <-ctx.Done():
		h.logger.Warn("Health check timeout", map[string]interface{}{
			"timeout": h.timeout,
		})
		return
	}

	close(results)

	// Collect results
	var components []ComponentHealth
	for result := range results {
		components = append(components, result)
	}

	// Generate report
	report := h.generateReport(components)

	// Store last report
	h.reportMu.Lock()
	h.lastReport = report
	h.reportMu.Unlock()

	// Log overall status
	h.logger.Info("Health check completed", map[string]interface{}{
		"overallStatus": report.OverallStatus,
		"components":    len(report.Components),
		"summary":       report.Summary,
	})
}

// generateReport generates a health report from component results
func (h *HealthChecker) generateReport(components []ComponentHealth) HealthReport {
	report := HealthReport{
		Timestamp:  time.Now(),
		Components: components,
		Summary:    make(map[string]int),
		Uptime:     time.Since(h.startTime),
		Data:       make(map[string]interface{}),
	}

	// Count statuses
	for _, component := range components {
		report.Summary[string(component.Status)]++
	}

	// Determine overall status
	report.OverallStatus = h.determineOverallStatus(components)

	// Add system information
	report.Data["system"] = h.getSystemInfo()

	return report
}

// determineOverallStatus determines the overall health status
func (h *HealthChecker) determineOverallStatus(components []ComponentHealth) HealthStatus {
	if len(components) == 0 {
		return HealthStatusUnknown
	}

	unhealthy := 0
	degraded := 0

	for _, component := range components {
		switch component.Status {
		case HealthStatusUnhealthy:
			unhealthy++
		case HealthStatusDegraded:
			degraded++
		}
	}

	if unhealthy > 0 {
		return HealthStatusUnhealthy
	}
	if degraded > 0 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// getSystemInfo returns system information
func (h *HealthChecker) getSystemInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"num_gc":      m.NumGC,
		},
		"cpu_count": runtime.NumCPU(),
	}
}

// GetLastReport returns the last health report
func (h *HealthChecker) GetLastReport() HealthReport {
	h.reportMu.RLock()
	defer h.reportMu.RUnlock()
	return h.lastReport
}

// GetStatus returns the current status
func (h *HealthChecker) GetStatus() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	report := h.GetLastReport()

	status := map[string]interface{}{
		"running":       h.running,
		"startTime":     h.startTime,
		"checkInterval": h.checkInterval,
		"timeout":       h.timeout,
		"checks":        len(h.checks),
		"overallStatus": report.OverallStatus,
		"uptime":        report.Uptime,
	}

	return status
}

// ForceCheck forces an immediate health check
func (h *HealthChecker) ForceCheck() HealthReport {
	// Create a temporary context if health checker is not running
	ctx := h.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	h.mu.RLock()
	checks := make(map[string]HealthCheck)
	for k, v := range h.checks {
		checks[k] = v
	}
	h.mu.RUnlock()

	if len(checks) == 0 {
		return h.generateReport([]ComponentHealth{})
	}

	// Limit concurrent checks to prevent DoS
	maxConcurrent := 10
	if len(checks) > maxConcurrent {
		h.logger.Warn("Too many health checks, limiting concurrent execution", map[string]interface{}{
			"total":         len(checks),
			"maxConcurrent": maxConcurrent,
		})
	}

	// Run checks with concurrency limit
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	results := make(chan ComponentHealth, len(checks))

	for name, check := range checks {
		wg.Add(1)
		go func(c HealthCheck, n string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-checkCtx.Done():
				return
			}

			// Run check with individual timeout
			individualCtx, individualCancel := context.WithTimeout(checkCtx, h.timeout/2)
			defer individualCancel()

			result := c.Check(individualCtx)
			select {
			case results <- result:
			case <-checkCtx.Done():
			}
		}(check, name)
	}

	// Wait for all checks to complete with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// All checks completed
	case <-checkCtx.Done():
		h.logger.Warn("Health check timeout", map[string]interface{}{
			"timeout": h.timeout,
		})
		return h.generateReport([]ComponentHealth{})
	}

	close(results)

	// Collect results
	var components []ComponentHealth
	for result := range results {
		components = append(components, result)
	}

	// Generate and return report
	return h.generateReport(components)
}
