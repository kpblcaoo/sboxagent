package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/kpblcaoo/sboxagent/internal/utils"
)

// CLIRequest represents a CLI request to sboxmgr
type CLIRequest struct {
	RequestID       string                 `json:"request_id"`
	Timestamp       string                 `json:"timestamp"`
	ProtocolVersion string                 `json:"protocol_version"`
	Action          string                 `json:"action"`
	SubscriptionURL string                 `json:"subscription_url,omitempty"`
	ClientType      string                 `json:"client_type,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

// CLIResponse represents a CLI response from sboxmgr
type CLIResponse struct {
	RequestID       string                 `json:"request_id"`
	Timestamp       string                 `json:"timestamp"`
	ProtocolVersion string                 `json:"protocol_version"`
	Success         bool                   `json:"success"`
	Data            map[string]interface{} `json:"data,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// CLIService represents the CLI integration service
type CLIService struct {
	config config.CLIConfig
	logger *logger.Logger

	// State
	mu        sync.RWMutex
	running   bool
	lastRun   time.Time
	lastError error

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCLIService creates a new CLI service
func NewCLIService(cfg config.CLIConfig, log *logger.Logger) (*CLIService, error) {
	return &CLIService{
		config: cfg,
		logger: log,
	}, nil
}

// Start starts the CLI service
func (s *CLIService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("CLI service is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	s.logger.Info("CLI service starting", map[string]interface{}{
		"sboxmgr_path": s.config.SboxmgrPath,
		"timeout":      s.config.Timeout,
		"max_retries":  s.config.MaxRetries,
	})

	return nil
}

// Stop stops the CLI service
func (s *CLIService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("CLI service stopping", map[string]interface{}{})
	s.cancel()
	s.running = false
}

// ExecuteCommand executes a sboxmgr command
func (s *CLIService) ExecuteCommand(args []string) ([]byte, error) {
	s.mu.Lock()
	s.lastRun = time.Now()
	s.mu.Unlock()

	s.logger.Debug("Executing sboxmgr command", map[string]interface{}{
		"command": args,
	})

	// Parse timeout
	timeout, err := utils.ParseDuration(s.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout format: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	// Build command
	cmd := exec.CommandContext(ctx, s.config.SboxmgrPath, args...)

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		s.setLastError(err)
		return nil, fmt.Errorf("failed to execute sboxmgr command: %w", err)
	}

	s.setLastError(nil)
	return output, nil
}

// GenerateConfig generates a configuration using sboxmgr
func (s *CLIService) GenerateConfig(subscriptionURL, clientType string, options map[string]interface{}) (*CLIResponse, error) {
	// Build command arguments
	args := []string{"json", "generate", "-u", subscriptionURL, "-c", clientType}

	// Add options
	if excludeList, ok := options["exclude"].(string); ok && excludeList != "" {
		args = append(args, "--exclude", excludeList)
	}

	if includeList, ok := options["include"].(string); ok && includeList != "" {
		args = append(args, "--include", includeList)
	}

	if version, ok := options["version"].(string); ok && version != "" {
		args = append(args, "--version", version)
	}

	// Execute command
	output, err := s.ExecuteCommand(args)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response CLIResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sboxmgr response: %w", err)
	}

	s.logger.Info("Configuration generated successfully", map[string]interface{}{
		"client": clientType,
		"url":    subscriptionURL,
	})

	return &response, nil
}

// ValidateConfig validates a configuration using sboxmgr
func (s *CLIService) ValidateConfig(configPath, clientType string) (*CLIResponse, error) {
	// Build command arguments
	args := []string{"json", "validate", "-f", configPath, "-c", clientType}

	// Execute command
	output, err := s.ExecuteCommand(args)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response CLIResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sboxmgr response: %w", err)
	}

	s.logger.Info("Configuration validation completed", map[string]interface{}{
		"config": configPath,
		"client": clientType,
		"valid":  response.Success,
	})

	return &response, nil
}

// ListClients lists available clients using sboxmgr
func (s *CLIService) ListClients() (*CLIResponse, error) {
	// Build command arguments
	args := []string{"json", "list-clients"}

	// Execute command
	output, err := s.ExecuteCommand(args)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response CLIResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sboxmgr response: %w", err)
	}

	s.logger.Info("Client list retrieved", map[string]interface{}{
		"clients": response.Data,
	})

	return &response, nil
}

// GetInfo gets sboxmgr information
func (s *CLIService) GetInfo() (*CLIResponse, error) {
	// Build command arguments
	args := []string{"json", "info"}

	// Execute command
	output, err := s.ExecuteCommand(args)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response CLIResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sboxmgr response: %w", err)
	}

	s.logger.Info("Sboxmgr info retrieved", map[string]interface{}{
		"info": response.Data,
	})

	return &response, nil
}

// ExecuteWithRetry executes a command with retry logic
func (s *CLIService) ExecuteWithRetry(args []string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Debug("Retrying command", map[string]interface{}{
				"attempt": attempt,
				"command": args,
			})

			// Wait before retry
			retryInterval, err := utils.ParseDuration(s.config.RetryInterval)
			if err != nil {
				return nil, fmt.Errorf("invalid retry interval format: %w", err)
			}
			time.Sleep(retryInterval)
		}

		output, err := s.ExecuteCommand(args)
		if err == nil {
			return output, nil
		}

		lastErr = err
		s.logger.Warn("Command execution failed", map[string]interface{}{
			"attempt": attempt,
			"error":   err.Error(),
		})
	}

	return nil, fmt.Errorf("command failed after %d attempts: %w", s.config.MaxRetries+1, lastErr)
}

// setLastError sets the last error
func (s *CLIService) setLastError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = err
}

// GetStatus returns the current service status
func (s *CLIService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"running":      s.running,
		"lastRun":      s.lastRun,
		"sboxmgr_path": s.config.SboxmgrPath,
		"timeout":      s.config.Timeout,
		"max_retries":  s.config.MaxRetries,
	}

	if s.lastError != nil {
		status["lastError"] = s.lastError.Error()
	}

	return status
}

// IsRunning returns true if the service is running
func (s *CLIService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
