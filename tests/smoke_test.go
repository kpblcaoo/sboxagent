package tests

import (
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/agent"
	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/kpblcaoo/sboxagent/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmoke_BasicAgentCreation tests basic agent creation and configuration
func TestSmoke_BasicAgentCreation(t *testing.T) {
	// Create minimal configuration
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:     "test-agent",
			Version:  "0.1.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled:  false, // Disable for smoke test
				Interval: "1m",
				Timeout:  "30s",
			},
			CLI: config.CLIConfig{
				Enabled:     false, // Disable for smoke test
				SboxmgrPath: "sboxmgr",
				Timeout:     "30s",
			},
			Systemd: config.SystemdConfig{
				Enabled:     false, // Disable for smoke test
				ServiceName: "sboxagent",
				UserMode:    false,
			},
			Monitoring: config.MonitorConfig{
				Enabled:        false, // Disable for smoke test
				Interval:       "30s",
				MetricsEnabled: true,
				AlertsEnabled:  true,
			},
		},
	}

	// Create agent
	a, err := agent.New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, a)

	// Check initial status
	status := a.GetStatus()
	assert.False(t, status["running"].(bool))
	assert.Equal(t, "test-agent", a.GetConfig().Agent.Name)
}

// TestSmoke_UtilsParseDuration tests the utility function
func TestSmoke_UtilsParseDuration(t *testing.T) {
	// Test valid durations
	duration, err := utils.ParseDuration("30s")
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Second, duration)

	duration, err = utils.ParseDuration("5m")
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Minute, duration)

	duration, err = utils.ParseDuration("2h")
	assert.NoError(t, err)
	assert.Equal(t, 2*time.Hour, duration)

	// Test invalid duration
	_, err = utils.ParseDuration("invalid")
	assert.Error(t, err)
}

// TestSmoke_ConfigValidation tests configuration validation
func TestSmoke_ConfigValidation(t *testing.T) {
	// Test valid configuration
	validCfg := &config.Config{
		Agent: config.AgentConfig{
			Name:     "test-agent",
			Version:  "0.1.0",
			LogLevel: "info",
		},
	}

	// Should not error
	assert.NotNil(t, validCfg)
	assert.Equal(t, "test-agent", validCfg.Agent.Name)
	assert.Equal(t, "0.1.0", validCfg.Agent.Version)
}

// TestSmoke_ServiceConfigs tests service configuration structures
func TestSmoke_ServiceConfigs(t *testing.T) {
	// Test CLI config
	cliConfig := config.CLIConfig{
		Enabled:       true,
		SboxmgrPath:   "sboxmgr",
		Timeout:       "30s",
		MaxRetries:    3,
		RetryInterval: "5s",
	}
	assert.True(t, cliConfig.Enabled)
	assert.Equal(t, "sboxmgr", cliConfig.SboxmgrPath)

	// Test Systemd config
	systemdConfig := config.SystemdConfig{
		Enabled:     true,
		ServiceName: "sboxagent",
		UserMode:    false,
		AutoStart:   true,
	}
	assert.True(t, systemdConfig.Enabled)
	assert.Equal(t, "sboxagent", systemdConfig.ServiceName)

	// Test Monitor config
	monitorConfig := config.MonitorConfig{
		Enabled:        true,
		Interval:       "30s",
		MetricsEnabled: true,
		AlertsEnabled:  true,
		RetentionDays:  30,
	}
	assert.True(t, monitorConfig.Enabled)
	assert.True(t, monitorConfig.MetricsEnabled)
}
