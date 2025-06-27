package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/logger"
)

// SboxctlEvent represents an event from sboxctl
type SboxctlEvent struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
}

// SboxctlService represents the sboxctl service
type SboxctlService struct {
	config config.SboxctlConfig
	logger *logger.Logger
	
	// State
	mu       sync.RWMutex
	running  bool
	lastRun  time.Time
	lastError error
	
	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	
	// Event handling
	eventChan chan SboxctlEvent
}

// NewSboxctlService creates a new sboxctl service
func NewSboxctlService(cfg config.SboxctlConfig, log *logger.Logger) (*SboxctlService, error) {
	return &SboxctlService{
		config:    cfg,
		logger:    log,
		eventChan: make(chan SboxctlEvent, 100), // Buffer for events
	}, nil
}

// Start starts the sboxctl service
func (s *SboxctlService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("sboxctl service is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	s.logger.Info("Sboxctl service starting", map[string]interface{}{
		"command":  s.config.Command,
		"interval": s.config.Interval,
		"timeout":  s.config.Timeout,
	})

	// Start the main service loop
	go s.run()

	// Start health checker if enabled
	if s.config.HealthCheck.Enabled {
		go s.healthChecker()
	}

	return nil
}

// Stop stops the sboxctl service
func (s *SboxctlService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Sboxctl service stopping", map[string]interface{}{})
	s.cancel()
	s.running = false
}

// run is the main service loop
func (s *SboxctlService) run() {
	// Parse interval
	interval, err := parseDuration(s.config.Interval)
	if err != nil {
		s.logger.Error("Invalid interval format", map[string]interface{}{
			"interval": s.config.Interval,
			"error":    err.Error(),
		})
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial execution
	s.executeSboxctl()

	// Main loop
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Sboxctl service loop stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			s.executeSboxctl()
		}
	}
}

// executeSboxctl executes the sboxctl command and captures output
func (s *SboxctlService) executeSboxctl() {
	s.mu.Lock()
	s.lastRun = time.Now()
	s.mu.Unlock()

	s.logger.Debug("Executing sboxctl command", map[string]interface{}{
		"command": s.config.Command,
	})

	// Parse timeout
	timeout, err := parseDuration(s.config.Timeout)
	if err != nil {
		s.logger.Error("Invalid timeout format", map[string]interface{}{
			"timeout": s.config.Timeout,
			"error":   err.Error(),
		})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, s.config.Command[0], s.config.Command[1:]...)

	// Capture stdout if enabled
	if s.config.StdoutCapture {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			s.logger.Error("Failed to create stdout pipe", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Start reading stdout in a goroutine
		go s.readStdout(stdout)
	}

	// Execute command
	if err := cmd.Start(); err != nil {
		s.logger.Error("Failed to start sboxctl command", map[string]interface{}{
			"command": s.config.Command,
			"error":   err.Error(),
		})
		s.setLastError(err)
		return
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		s.logger.Error("Sboxctl command failed", map[string]interface{}{
			"command": s.config.Command,
			"error":   err.Error(),
		})
		s.setLastError(err)
		return
	}

	s.logger.Info("Sboxctl command completed successfully", map[string]interface{}{
		"command": s.config.Command,
	})
	s.setLastError(nil)
}

// readStdout reads and processes stdout from sboxctl
func (s *SboxctlService) readStdout(stdout interface{}) {
	scanner := bufio.NewScanner(stdout.(interface{ Read(p []byte) (n int, err error) }))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		s.logger.Debug("Received stdout line", map[string]interface{}{
			"line": line,
		})

		// Try to parse as JSON event
		if event, err := s.parseEvent(line); err == nil {
			s.handleEvent(event)
		} else {
			// Treat as plain log line
			s.logger.Info("Sboxctl output", map[string]interface{}{
				"output": line,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("Error reading stdout", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// parseEvent attempts to parse a line as a JSON event
func (s *SboxctlService) parseEvent(line string) (*SboxctlEvent, error) {
	var event SboxctlEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return nil, err
	}

	// Validate event
	if event.Type == "" {
		return nil, fmt.Errorf("event type is required")
	}

	return &event, nil
}

// handleEvent processes a parsed event
func (s *SboxctlService) handleEvent(event *SboxctlEvent) {
	s.logger.Info("Processing sboxctl event", map[string]interface{}{
		"type":      event.Type,
		"timestamp": event.Timestamp,
		"version":   event.Version,
	})

	// Send event to channel for further processing
	select {
	case s.eventChan <- *event:
		// Event sent successfully
	default:
		// Channel is full, log warning
		s.logger.Warn("Event channel is full, dropping event", map[string]interface{}{
			"type": event.Type,
		})
	}
}

// healthChecker runs periodic health checks
func (s *SboxctlService) healthChecker() {
	interval, err := parseDuration(s.config.HealthCheck.Interval)
	if err != nil {
		s.logger.Error("Invalid health check interval", map[string]interface{}{
			"interval": s.config.HealthCheck.Interval,
			"error":    err.Error(),
		})
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performHealthCheck()
		}
	}
}

// performHealthCheck performs a health check
func (s *SboxctlService) performHealthCheck() {
	s.mu.RLock()
	lastRun := s.lastRun
	lastError := s.lastError
	s.mu.RUnlock()

	// Check if last run was successful
	if lastError != nil {
		s.logger.Warn("Health check failed", map[string]interface{}{
			"lastError": lastError.Error(),
			"lastRun":   lastRun,
		})
	} else {
		s.logger.Debug("Health check passed", map[string]interface{}{
			"lastRun": lastRun,
		})
	}
}

// setLastError sets the last error
func (s *SboxctlService) setLastError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = err
}

// GetStatus returns the current service status
func (s *SboxctlService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"running":   s.running,
		"lastRun":   s.lastRun,
		"command":   s.config.Command,
		"interval":  s.config.Interval,
		"timeout":   s.config.Timeout,
	}

	if s.lastError != nil {
		status["lastError"] = s.lastError.Error()
	}

	return status
}

// GetEventChannel returns the event channel
func (s *SboxctlService) GetEventChannel() <-chan SboxctlEvent {
	return s.eventChan
}

// parseDuration parses a duration string (e.g., "30m", "5m", "10s")
func parseDuration(duration string) (time.Duration, error) {
	// Handle common formats
	switch {
	case strings.HasSuffix(duration, "s"):
		val, err := strconv.Atoi(strings.TrimSuffix(duration, "s"))
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Second, nil
	case strings.HasSuffix(duration, "m"):
		val, err := strconv.Atoi(strings.TrimSuffix(duration, "m"))
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Minute, nil
	case strings.HasSuffix(duration, "h"):
		val, err := strconv.Atoi(strings.TrimSuffix(duration, "h"))
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Hour, nil
	default:
		return time.ParseDuration(duration)
	}
} 