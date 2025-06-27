package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/kpblcaoo/sboxagent/internal/services"
)

// Agent represents the main agent instance
type Agent struct {
	config *config.Config
	logger *logger.Logger
	
	// Services
	sboxctlService *services.SboxctlService
	
	// State
	mu       sync.RWMutex
	running  bool
	startTime time.Time
	
	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new agent instance
func New(cfg *config.Config) (*Agent, error) {
	// Create logger
	log, err := logger.New(cfg.Agent.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create agent
	agent := &Agent{
		config: cfg,
		logger: log,
	}

	// Initialize services
	if err := agent.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return agent, nil
}

// initializeServices initializes all agent services
func (a *Agent) initializeServices() error {
	// Initialize sboxctl service if enabled
	if a.config.Services.Sboxctl.Enabled {
		sboxctlService, err := services.NewSboxctlService(a.config.Services.Sboxctl, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create sboxctl service: %w", err)
		}
		a.sboxctlService = sboxctlService
	}

	return nil
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("agent is already running")
	}

	// Create context for graceful shutdown
	a.ctx, a.cancel = context.WithCancel(ctx)
	defer a.cancel()

	a.running = true
	a.startTime = time.Now()

	a.logger.Info("Agent starting", map[string]interface{}{
		"name":    a.config.Agent.Name,
		"version": a.config.Agent.Version,
	})

	// Start services
	if err := a.startServices(); err != nil {
		a.running = false
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for context cancellation
	<-a.ctx.Done()

	// Stop services
	a.stopServices()

	a.running = false
	a.logger.Info("Agent stopped", map[string]interface{}{})

	return nil
}

// startServices starts all enabled services
func (a *Agent) startServices() error {
	// Start sboxctl service
	if a.sboxctlService != nil {
		if err := a.sboxctlService.Start(a.ctx); err != nil {
			return fmt.Errorf("failed to start sboxctl service: %w", err)
		}
		a.logger.Info("Sboxctl service started", map[string]interface{}{})
	}

	return nil
}

// stopServices stops all running services
func (a *Agent) stopServices() {
	// Stop sboxctl service
	if a.sboxctlService != nil {
		a.sboxctlService.Stop()
		a.logger.Info("Sboxctl service stopped", map[string]interface{}{})
	}
}

// Stop stops the agent gracefully
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return
	}

	a.logger.Info("Stopping agent", map[string]interface{}{})
	a.cancel()
}

// IsRunning returns true if the agent is running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetStatus returns the current agent status
func (a *Agent) GetStatus() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := map[string]interface{}{
		"running":   a.running,
		"startTime": a.startTime,
		"uptime":    time.Since(a.startTime).String(),
	}

	if a.sboxctlService != nil {
		status["sboxctl"] = a.sboxctlService.GetStatus()
	}

	return status
}

// GetConfig returns the current configuration
func (a *Agent) GetConfig() *config.Config {
	return a.config
} 