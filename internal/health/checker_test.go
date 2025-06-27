package health

import (
	"context"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
)

func TestNewHealthChecker(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 30*time.Second, 5*time.Second)

	if checker == nil {
		t.Fatal("Expected health checker to be created")
	}

	if checker.logger != log {
		t.Error("Expected logger to be set")
	}

	if checker.checkInterval != 30*time.Second {
		t.Errorf("Expected checkInterval to be 30s, got %v", checker.checkInterval)
	}

	if checker.timeout != 5*time.Second {
		t.Errorf("Expected timeout to be 5s, got %v", checker.timeout)
	}
}

func TestHealthChecker_RegisterCheck(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 1*time.Second, 500*time.Millisecond)

	// Create a test check
	testCheck := &testHealthCheck{
		name: "test_check",
	}

	// Register check
	err := checker.RegisterCheck(testCheck)
	if err != nil {
		t.Fatalf("Expected no error on check registration, got: %v", err)
	}

	// Check status
	status := checker.GetStatus()
	if status["checks"].(int) != 1 {
		t.Error("Expected 1 registered check")
	}
}

func TestHealthChecker_RegisterNilCheck(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 1*time.Second, 500*time.Millisecond)

	err := checker.RegisterCheck(nil)
	if err == nil {
		t.Error("Expected error when registering nil check")
	}
}

func TestHealthChecker_UnregisterCheck(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 1*time.Second, 500*time.Millisecond)

	// Create and register check
	testCheck := &testHealthCheck{
		name: "test_check",
	}
	checker.RegisterCheck(testCheck)

	// Unregister check
	checker.UnregisterCheck("test_check")

	// Check status
	status := checker.GetStatus()
	if status["checks"].(int) != 0 {
		t.Error("Expected 0 registered checks after unregister")
	}
}

func TestHealthChecker_ForceCheck(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 1*time.Second, 500*time.Millisecond)

	// Create and register a test check (without starting the checker)
	testCheck := &testHealthCheck{
		name:   "test_check",
		status: HealthStatusHealthy,
	}
	checker.RegisterCheck(testCheck)

	// Force a health check
	report := checker.ForceCheck()

	// Check report
	if report.OverallStatus != HealthStatusHealthy {
		t.Errorf("Expected overall status to be healthy, got %s", report.OverallStatus)
	}

	if len(report.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(report.Components))
	}

	if report.Components[0].Name != "test_check" {
		t.Errorf("Expected component name to be 'test_check', got %s", report.Components[0].Name)
	}
}

func TestHealthChecker_DetermineOverallStatus(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 1*time.Second, 500*time.Millisecond)

	// Test with no components
	components := []ComponentHealth{}
	status := checker.determineOverallStatus(components)
	if status != HealthStatusUnknown {
		t.Errorf("Expected unknown status for no components, got %s", status)
	}

	// Test with healthy components
	components = []ComponentHealth{
		{Status: HealthStatusHealthy},
		{Status: HealthStatusHealthy},
	}
	status = checker.determineOverallStatus(components)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", status)
	}

	// Test with degraded components
	components = []ComponentHealth{
		{Status: HealthStatusHealthy},
		{Status: HealthStatusDegraded},
	}
	status = checker.determineOverallStatus(components)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status, got %s", status)
	}

	// Test with unhealthy components
	components = []ComponentHealth{
		{Status: HealthStatusHealthy},
		{Status: HealthStatusDegraded},
		{Status: HealthStatusUnhealthy},
	}
	status = checker.determineOverallStatus(components)
	if status != HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", status)
	}
}

func TestHealthChecker_GetSystemInfo(t *testing.T) {
	log, _ := logger.New("debug")
	checker := NewHealthChecker(log, 1*time.Second, 500*time.Millisecond)

	systemInfo := checker.getSystemInfo()

	// Check that system info contains expected fields
	if systemInfo["goroutines"] == nil {
		t.Error("Expected goroutines field in system info")
	}

	if systemInfo["memory"] == nil {
		t.Error("Expected memory field in system info")
	}

	if systemInfo["cpu_count"] == nil {
		t.Error("Expected cpu_count field in system info")
	}

	// Check memory info structure
	memory, ok := systemInfo["memory"].(map[string]interface{})
	if !ok {
		t.Error("Expected memory to be a map")
	}

	if memory["alloc"] == nil {
		t.Error("Expected alloc field in memory info")
	}

	if memory["sys"] == nil {
		t.Error("Expected sys field in memory info")
	}
}

// testHealthCheck is a test implementation of HealthCheck
type testHealthCheck struct {
	name   string
	status HealthStatus
}

func (h *testHealthCheck) Name() string {
	return h.name
}

func (h *testHealthCheck) Check(ctx context.Context) ComponentHealth {
	return ComponentHealth{
		Name:      h.name,
		Status:    h.status,
		Message:   "Test check completed",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"test": true,
		},
	}
}
