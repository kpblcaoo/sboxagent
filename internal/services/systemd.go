package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/logger"
)

// SystemdService represents the systemd integration service
type SystemdService struct {
	logger *logger.Logger

	// Configuration
	serviceName string
	userMode    bool

	// State
	mu      sync.RWMutex
	running bool
	enabled bool
	status  string

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSystemdService creates a new systemd service
func NewSystemdService(cfg config.SystemdConfig, log *logger.Logger) (*SystemdService, error) {
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "sboxagent"
	}

	return &SystemdService{
		logger:      log,
		serviceName: serviceName,
		userMode:    cfg.UserMode,
	}, nil
}

// Start starts the systemd service
func (s *SystemdService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("systemd service is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	s.logger.Info("Systemd service starting", map[string]interface{}{
		"service_name": s.serviceName,
		"user_mode":    s.userMode,
	})

	// Check if systemd is available
	if !s.isSystemdAvailable() {
		s.logger.Warn("Systemd not available, running in standalone mode", map[string]interface{}{})
		return nil
	}

	// Check service status
	if err := s.checkServiceStatus(); err != nil {
		s.logger.Warn("Failed to check service status", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}

// Stop stops the systemd service
func (s *SystemdService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Systemd service stopping", map[string]interface{}{})
	s.cancel()
	s.running = false
}

// EnableService enables the systemd service
func (s *SystemdService) EnableService() error {
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd not available")
	}

	args := []string{"enable", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	s.mu.Lock()
	s.enabled = true
	s.mu.Unlock()

	s.logger.Info("Systemd service enabled", map[string]interface{}{
		"service_name": s.serviceName,
	})

	return nil
}

// DisableService disables the systemd service
func (s *SystemdService) DisableService() error {
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd not available")
	}

	args := []string{"disable", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}

	s.mu.Lock()
	s.enabled = false
	s.mu.Unlock()

	s.logger.Info("Systemd service disabled", map[string]interface{}{
		"service_name": s.serviceName,
	})

	return nil
}

// StartService starts the systemd service
func (s *SystemdService) StartService() error {
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd not available")
	}

	args := []string{"start", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	s.logger.Info("Systemd service started", map[string]interface{}{
		"service_name": s.serviceName,
	})

	return nil
}

// StopService stops the systemd service
func (s *SystemdService) StopService() error {
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd not available")
	}

	args := []string{"stop", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	s.logger.Info("Systemd service stopped", map[string]interface{}{
		"service_name": s.serviceName,
	})

	return nil
}

// RestartService restarts the systemd service
func (s *SystemdService) RestartService() error {
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd not available")
	}

	args := []string{"restart", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}

	s.logger.Info("Systemd service restarted", map[string]interface{}{
		"service_name": s.serviceName,
	})

	return nil
}

// ReloadService reloads the systemd service
func (s *SystemdService) ReloadService() error {
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd not available")
	}

	args := []string{"reload", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload service: %w", err)
	}

	s.logger.Info("Systemd service reloaded", map[string]interface{}{
		"service_name": s.serviceName,
	})

	return nil
}

// GetServiceStatus gets the current service status
func (s *SystemdService) GetServiceStatus() (string, error) {
	if !s.isSystemdAvailable() {
		return "systemd-not-available", nil
	}

	args := []string{"is-active", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "unknown", fmt.Errorf("failed to get service status: %w", err)
	}

	status := strings.TrimSpace(string(output))
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()

	return status, nil
}

// IsServiceEnabled checks if the service is enabled
func (s *SystemdService) IsServiceEnabled() (bool, error) {
	if !s.isSystemdAvailable() {
		return false, fmt.Errorf("systemd not available")
	}

	args := []string{"is-enabled", s.serviceName}
	if s.userMode {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return false, nil
	}

	return true, nil
}

// checkServiceStatus checks the current service status
func (s *SystemdService) checkServiceStatus() error {
	status, err := s.GetServiceStatus()
	if err != nil {
		return err
	}

	enabled, err := s.IsServiceEnabled()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.status = status
	s.enabled = enabled
	s.mu.Unlock()

	s.logger.Info("Service status checked", map[string]interface{}{
		"service_name": s.serviceName,
		"status":       status,
		"enabled":      enabled,
	})

	return nil
}

// isSystemdAvailable checks if systemd is available
func (s *SystemdService) isSystemdAvailable() bool {
	// Check if systemctl is available
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false
	}

	// Check if systemd is running
	cmd := exec.Command("systemctl", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	// Check for systemd runtime directory
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetStatus returns the current service status
func (s *SystemdService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"running":      s.running,
		"service_name": s.serviceName,
		"user_mode":    s.userMode,
		"enabled":      s.enabled,
		"status":       s.status,
	}

	// Check if systemd is available
	status["systemd_available"] = s.isSystemdAvailable()

	return status
}

// IsRunning returns true if the service is running
func (s *SystemdService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// CreateServiceFile creates a systemd service file
func (s *SystemdService) CreateServiceFile(execPath, configPath string) error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=SboxAgent - Subbox Management Agent
After=network.target

[Service]
Type=simple
User=sboxagent
Group=sboxagent
ExecStart=%s --config %s
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, execPath, configPath)

	// Determine service file path
	var servicePath string
	if s.userMode {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		serviceDir := filepath.Join(homeDir, ".config/systemd/user")
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			return fmt.Errorf("failed to create service directory: %w", err)
		}
		servicePath = filepath.Join(serviceDir, s.serviceName+".service")
	} else {
		servicePath = filepath.Join("/etc/systemd/system", s.serviceName+".service")
	}

	// Write service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	s.logger.Info("Systemd service file created", map[string]interface{}{
		"service_path": servicePath,
		"user_mode":    s.userMode,
	})

	return nil
}
