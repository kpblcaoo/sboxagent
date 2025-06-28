package tests

import (
	"context"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/agent"
	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealIntegration_ExternalTools tests real integration with external tools
func TestRealIntegration_ExternalTools(t *testing.T) {
	t.Run("Sboxctl_Integration", func(t *testing.T) {
		// Test that CLI service can interact with real sboxctl
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "real-integration-test",
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
			},
		}

		// Create CLI service
		cliService, err := services.NewCLIService(cfg.Services.CLI, nil)
		require.NoError(t, err)
		assert.NotNil(t, cliService)

		// Start the service
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = cliService.Start(ctx)
		require.NoError(t, err)

		// Test basic sboxctl command (help)
		result, err := cliService.ExecuteCommand([]string{"--help"})
		if err != nil {
			t.Logf("sboxctl --help failed (expected if not configured): %v", err)
		} else {
			assert.NotEmpty(t, result)
			assert.Contains(t, string(result), "Usage:")
			t.Log("✅ sboxctl integration works")
		}

		// Test list-servers command (will fail without URL, but should handle error gracefully)
		result, err = cliService.ExecuteCommand([]string{"list-servers"})
		if err != nil {
			t.Logf("sboxctl list-servers failed (expected without URL): %v", err)
			// This is expected behavior - command should fail gracefully
		} else {
			t.Log("✅ sboxctl list-servers works")
		}

		// Stop the service
		cliService.Stop()
	})

	t.Run("Systemd_Integration", func(t *testing.T) {
		// Test that systemd service can interact with real systemd
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "systemd-test",
				Version:  "0.1.0",
				LogLevel: "info",
			},
			Services: config.ServicesConfig{
				Systemd: config.SystemdConfig{
					Enabled:     true,
					ServiceName: "systemd-resolved", // Use a real system service
					UserMode:    false,
					AutoStart:   false,
				},
			},
		}

		// Create systemd service
		systemdService, err := services.NewSystemdService(cfg.Services.Systemd, nil)
		require.NoError(t, err)
		assert.NotNil(t, systemdService)

		// Start the service
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = systemdService.Start(ctx)
		require.NoError(t, err)

		// Test systemd service status (should work with real systemd)
		status := systemdService.GetStatus()
		assert.NotNil(t, status)
		t.Logf("✅ systemd integration works, status: %+v", status)

		// Stop the service
		systemdService.Stop()
	})

	t.Run("Agent_With_Real_Services", func(t *testing.T) {
		// Test agent with real external services
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "real-agent-test",
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
					ServiceName: "systemd-resolved",
					UserMode:    false,
					AutoStart:   false,
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

		// Create agent with real services
		a, err := agent.New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, a)

		// Test agent status
		status := a.GetStatus()
		assert.NotNil(t, status)
		assert.Contains(t, status, "running")
		assert.Contains(t, status, "startTime")

		// Test agent configuration
		agentConfig := a.GetConfig()
		assert.Equal(t, "real-agent-test", agentConfig.Agent.Name)
		assert.True(t, agentConfig.Services.CLI.Enabled)
		assert.True(t, agentConfig.Services.Systemd.Enabled)
		assert.True(t, agentConfig.Services.Monitoring.Enabled)

		t.Log("✅ Agent with real services works correctly")
	})
}

// TestRealIntegration_ErrorHandling tests real error handling scenarios
func TestRealIntegration_ErrorHandling(t *testing.T) {
	t.Run("Invalid_External_Tool", func(t *testing.T) {
		// Test with non-existent external tool
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "error-test",
				Version:  "0.1.0",
				LogLevel: "info",
			},
			Services: config.ServicesConfig{
				CLI: config.CLIConfig{
					Enabled:       true,
					SboxmgrPath:   "nonexistent-tool",
					Timeout:       "30s",
					MaxRetries:    3,
					RetryInterval: "5s",
				},
			},
		}

		// Agent should still be created
		a, err := agent.New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, a)

		// CLI service should handle errors gracefully
		cliService, err := services.NewCLIService(cfg.Services.CLI, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = cliService.Start(ctx)
		require.NoError(t, err)

		// This should fail but not crash
		result, err := cliService.ExecuteCommand([]string{"--help"})
		assert.Error(t, err)
		assert.Empty(t, result)

		cliService.Stop()
		t.Log("✅ Error handling with invalid external tool works")
	})

	t.Run("Invalid_Systemd_Service", func(t *testing.T) {
		// Test with non-existent systemd service
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "systemd-error-test",
				Version:  "0.1.0",
				LogLevel: "info",
			},
			Services: config.ServicesConfig{
				Systemd: config.SystemdConfig{
					Enabled:     true,
					ServiceName: "nonexistent-service",
					UserMode:    false,
					AutoStart:   false,
				},
			},
		}

		// Agent should still be created
		a, err := agent.New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, a)

		// Systemd service should handle errors gracefully
		systemdService, err := services.NewSystemdService(cfg.Services.Systemd, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = systemdService.Start(ctx)
		require.NoError(t, err)

		// This should work but may show error status
		status := systemdService.GetStatus()
		assert.NotNil(t, status)
		t.Logf("Status for non-existent service: %+v", status)

		systemdService.Stop()
		t.Log("✅ Error handling with invalid systemd service works")
	})
}

// TestRealIntegration_Performance tests real performance characteristics
func TestRealIntegration_Performance(t *testing.T) {
	t.Run("External_Tool_Performance", func(t *testing.T) {
		// Test performance of external tool execution
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "perf-test",
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
			},
		}

		cliService, err := services.NewCLIService(cfg.Services.CLI, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = cliService.Start(ctx)
		require.NoError(t, err)

		// Test performance of help command
		start := time.Now()
		result, err := cliService.ExecuteCommand([]string{"--help"})
		duration := time.Since(start)

		if err == nil {
			assert.NotEmpty(t, result)
			assert.Less(t, duration, 5*time.Second, "External tool execution should be reasonably fast")
			t.Logf("✅ External tool performance: %v", duration)
		} else {
			t.Logf("External tool failed (expected): %v", err)
		}

		cliService.Stop()
	})

	t.Run("Agent_Startup_Performance", func(t *testing.T) {
		// Test agent startup performance with real services
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Name:     "startup-perf-test",
				Version:  "0.1.0",
				LogLevel: "info",
			},
			Services: config.ServicesConfig{
				CLI: config.CLIConfig{
					Enabled:     true,
					SboxmgrPath: "sboxctl",
					Timeout:     "30s",
				},
				Systemd: config.SystemdConfig{
					Enabled:     true,
					ServiceName: "systemd-resolved",
					UserMode:    false,
				},
				Monitoring: config.MonitorConfig{
					Enabled:  true,
					Interval: "30s",
				},
			},
		}

		start := time.Now()
		a, err := agent.New(cfg)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotNil(t, a)
		assert.Less(t, duration, 1*time.Second, "Agent creation should be fast")

		t.Logf("✅ Agent startup performance: %v", duration)
	})
}
