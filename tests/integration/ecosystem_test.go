package integration

import (
	"context"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/agent"
	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/services"
	"github.com/kpblcaoo/sboxagent/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEcosystem_CompleteIntegration tests the complete ecosystem integration
func TestEcosystem_CompleteIntegration(t *testing.T) {
	// Create a comprehensive configuration
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:     "test-ecosystem",
			Version:  "0.1.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled:       true,
				Command:       []string{"echo", "test"},
				Interval:      "1m",
				Timeout:       "30s",
				StdoutCapture: true,
				HealthCheck: config.HealthCheckConfig{
					Enabled:  true,
					Interval: "30s",
					Timeout:  "5s",
				},
			},
			CLI: config.CLIConfig{
				Enabled:       true,
				SboxmgrPath:   "sboxctl", // Use sboxctl as sboxmgr for testing
				Timeout:       "30s",
				MaxRetries:    3,
				RetryInterval: "5s",
			},
			Systemd: config.SystemdConfig{
				Enabled:     true,
				ServiceName: "test-service",
				UserMode:    false,
				AutoStart:   true,
			},
			Monitoring: config.MonitorConfig{
				Enabled:        true,
				Interval:       "30s",
				MetricsEnabled: true,
				AlertsEnabled:  true,
				RetentionDays:  30,
			},
		},
	}

	// Test 1: Configuration validation
	t.Run("Configuration_Validation", func(t *testing.T) {
		assert.NotNil(t, cfg)
		assert.Equal(t, "test-ecosystem", cfg.Agent.Name)
		assert.True(t, cfg.Services.CLI.Enabled)
		assert.True(t, cfg.Services.Systemd.Enabled)
		assert.True(t, cfg.Services.Monitoring.Enabled)
	})

	// Test 2: Agent creation
	t.Run("Agent_Creation", func(t *testing.T) {
		a, err := agent.New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, a)
		assert.False(t, a.IsRunning())

		// Test agent status
		status := a.GetStatus()
		assert.NotNil(t, status)
		assert.False(t, status["running"].(bool))
	})

	// Test 3: Service creation and initialization
	t.Run("Service_Creation", func(t *testing.T) {
		// Test CLI service creation
		cliService, err := services.NewCLIService(cfg.Services.CLI, nil)
		require.NoError(t, err)
		assert.NotNil(t, cliService)

		// Test Systemd service creation
		systemdService, err := services.NewSystemdService(cfg.Services.Systemd, nil)
		require.NoError(t, err)
		assert.NotNil(t, systemdService)

		// Test Monitor service creation
		monitorService, err := services.NewMonitorService(cfg, nil)
		require.NoError(t, err)
		assert.NotNil(t, monitorService)
	})

	// Test 4: Utility functions
	t.Run("Utility_Functions", func(t *testing.T) {
		// Test duration parsing
		duration, err := utils.ParseDuration("30s")
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, duration)

		duration, err = utils.ParseDuration("5m")
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Minute, duration)

		// Test invalid duration
		_, err = utils.ParseDuration("invalid")
		assert.Error(t, err)
	})

	// Test 5: Agent lifecycle
	t.Run("Agent_Lifecycle", func(t *testing.T) {
		a, err := agent.New(cfg)
		require.NoError(t, err)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Start agent in goroutine
		go func() {
			err := a.Start(ctx)
			assert.NoError(t, err)
		}()

		// Wait for agent to start
		time.Sleep(1 * time.Second)

		// Check that agent is running
		assert.True(t, a.IsRunning())

		// Get status
		status := a.GetStatus()
		assert.NotNil(t, status)
		assert.True(t, status["running"].(bool))

		// Stop agent
		a.Stop()

		// Wait for agent to stop
		time.Sleep(1 * time.Second)

		// Check that agent is stopped
		assert.False(t, a.IsRunning())
	})

	// Test 6: Service integration
	t.Run("Service_Integration", func(t *testing.T) {
		a, err := agent.New(cfg)
		require.NoError(t, err)

		// Test that services are properly initialized
		agentConfig := a.GetConfig()
		assert.NotNil(t, agentConfig)
		assert.Equal(t, cfg, agentConfig)

		// Test that all required services are configured
		assert.True(t, agentConfig.Services.CLI.Enabled)
		assert.True(t, agentConfig.Services.Systemd.Enabled)
		assert.True(t, agentConfig.Services.Monitoring.Enabled)
	})
}

// TestEcosystem_ServiceCommunication tests service communication
func TestEcosystem_ServiceCommunication(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:     "test-communication",
			Version:  "0.1.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			CLI: config.CLIConfig{
				Enabled:       true,
				SboxmgrPath:   "sboxctl",
				Timeout:       "30s",
				MaxRetries:    3,
				RetryInterval: "5s",
			},
			Systemd: config.SystemdConfig{
				Enabled:     true,
				ServiceName: "test-service",
				UserMode:    false,
			},
			Monitoring: config.MonitorConfig{
				Enabled:        true,
				Interval:       "30s",
				MetricsEnabled: true,
				AlertsEnabled:  true,
			},
		},
	}

	// Test service status reporting
	t.Run("Service_Status_Reporting", func(t *testing.T) {
		a, err := agent.New(cfg)
		require.NoError(t, err)

		// Get initial status
		status := a.GetStatus()
		assert.NotNil(t, status)
		assert.Contains(t, status, "running")
		assert.Contains(t, status, "startTime")
		assert.Contains(t, status, "uptime")
	})

	// Test configuration consistency
	t.Run("Configuration_Consistency", func(t *testing.T) {
		a, err := agent.New(cfg)
		require.NoError(t, err)

		// Verify configuration is consistent
		agentConfig := a.GetConfig()
		assert.Equal(t, cfg.Agent.Name, agentConfig.Agent.Name)
		assert.Equal(t, cfg.Services.CLI.Enabled, agentConfig.Services.CLI.Enabled)
		assert.Equal(t, cfg.Services.Systemd.Enabled, agentConfig.Services.Systemd.Enabled)
		assert.Equal(t, cfg.Services.Monitoring.Enabled, agentConfig.Services.Monitoring.Enabled)
	})
}

// TestEcosystem_ErrorHandling tests error handling across the ecosystem
func TestEcosystem_ErrorHandling(t *testing.T) {
	// Test with invalid configuration
	t.Run("Invalid_Configuration", func(t *testing.T) {
		invalidCfg := &config.Config{
			Agent: config.AgentConfig{
				Name:    "", // Invalid: empty name
				Version: "", // Invalid: empty version
			},
		}

		// Agent should still be created (validation happens in Load())
		a, err := agent.New(invalidCfg)
		require.NoError(t, err)
		assert.NotNil(t, a)
	})

	// Test with disabled services
	t.Run("Disabled_Services", func(t *testing.T) {
		disabledCfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "test-disabled",
				Version:  "0.1.0",
				LogLevel: "info",
			},
			Services: config.ServicesConfig{
				CLI: config.CLIConfig{
					Enabled: false, // Disabled
				},
				Systemd: config.SystemdConfig{
					Enabled: false, // Disabled
				},
				Monitoring: config.MonitorConfig{
					Enabled: false, // Disabled
				},
			},
		}

		a, err := agent.New(disabledCfg)
		require.NoError(t, err)
		assert.NotNil(t, a)

		// Agent should work even with disabled services
		status := a.GetStatus()
		assert.NotNil(t, status)
		assert.False(t, status["running"].(bool))
	})
}

// TestEcosystem_Performance tests basic performance characteristics
func TestEcosystem_Performance(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:     "test-performance",
			Version:  "0.1.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			CLI: config.CLIConfig{
				Enabled:       true,
				SboxmgrPath:   "sboxctl",
				Timeout:       "30s",
				MaxRetries:    3,
				RetryInterval: "5s",
			},
			Systemd: config.SystemdConfig{
				Enabled:     true,
				ServiceName: "test-service",
				UserMode:    false,
			},
			Monitoring: config.MonitorConfig{
				Enabled:        true,
				Interval:       "30s",
				MetricsEnabled: true,
				AlertsEnabled:  true,
			},
		},
	}

	// Test agent creation performance
	t.Run("Agent_Creation_Performance", func(t *testing.T) {
		start := time.Now()

		a, err := agent.New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, a)

		duration := time.Since(start)
		assert.Less(t, duration, 100*time.Millisecond, "Agent creation should be fast")
	})

	// Test status retrieval performance
	t.Run("Status_Retrieval_Performance", func(t *testing.T) {
		a, err := agent.New(cfg)
		require.NoError(t, err)

		start := time.Now()

		status := a.GetStatus()
		assert.NotNil(t, status)

		duration := time.Since(start)
		assert.Less(t, duration, 10*time.Millisecond, "Status retrieval should be very fast")
	})
}
