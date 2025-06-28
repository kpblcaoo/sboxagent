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
	cliService     *services.CLIService
	systemdService *services.SystemdService
	monitorService *services.MonitorService

	// State
	mu        sync.RWMutex
	running   bool
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

	// Initialize CLI service if enabled
	if a.config.Services.CLI.Enabled {
		cliService, err := services.NewCLIService(a.config.Services.CLI, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create CLI service: %w", err)
		}
		a.cliService = cliService
	}

	// Initialize systemd service if enabled
	if a.config.Services.Systemd.Enabled {
		systemdService, err := services.NewSystemdService(a.config.Services.Systemd, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create systemd service: %w", err)
		}
		a.systemdService = systemdService
	}

	// Initialize monitor service if enabled
	if a.config.Services.Monitoring.Enabled {
		monitorService, err := services.NewMonitorService(a.config, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create monitor service: %w", err)
		}
		a.monitorService = monitorService
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

	// Start CLI service
	if a.cliService != nil {
		if err := a.cliService.Start(a.ctx); err != nil {
			return fmt.Errorf("failed to start CLI service: %w", err)
		}
		a.logger.Info("CLI service started", map[string]interface{}{})
	}

	// Start systemd service
	if a.systemdService != nil {
		if err := a.systemdService.Start(a.ctx); err != nil {
			return fmt.Errorf("failed to start systemd service: %w", err)
		}
		a.logger.Info("Systemd service started", map[string]interface{}{})
	}

	// Start monitor service
	if a.monitorService != nil {
		if err := a.monitorService.Start(a.ctx); err != nil {
			return fmt.Errorf("failed to start monitor service: %w", err)
		}
		a.logger.Info("Monitor service started", map[string]interface{}{})
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

	// Stop CLI service
	if a.cliService != nil {
		a.cliService.Stop()
		a.logger.Info("CLI service stopped", map[string]interface{}{})
	}

	// Stop systemd service
	if a.systemdService != nil {
		a.systemdService.Stop()
		a.logger.Info("Systemd service stopped", map[string]interface{}{})
	}

	// Stop monitor service
	if a.monitorService != nil {
		a.monitorService.Stop()
		a.logger.Info("Monitor service stopped", map[string]interface{}{})
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

	if a.cliService != nil {
		status["cli"] = a.cliService.GetStatus()
	}

	if a.systemdService != nil {
		status["systemd"] = a.systemdService.GetStatus()
	}

	if a.monitorService != nil {
		status["monitor"] = a.monitorService.GetStatus()
	}

	return status
}

// GetConfig returns the current configuration
func (a *Agent) GetConfig() *config.Config {
	return a.config
}
