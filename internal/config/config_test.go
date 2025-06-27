package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithValidConfig(t *testing.T) {
	// Create temporary config file
	configContent := `
agent:
  name: "test-agent"
  version: "1.0.0"
  log_level: "debug"

server:
  port: 9090
  host: "localhost"
  timeout: "60s"

services:
  sboxctl:
    enabled: true
    command: ["sboxctl", "update", "--test"]
    interval: "15m"
    timeout: "2m"
    stdout_capture: true
    health_check:
      enabled: true
      interval: "30s"
      timeout: "5s"
`

	tmpFile, err := os.CreateTemp("", "agent_test_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := Load(tmpFile.Name())
	require.NoError(t, err)

	// Assert values
	assert.Equal(t, "test-agent", cfg.Agent.Name)
	assert.Equal(t, "1.0.0", cfg.Agent.Version)
	assert.Equal(t, "debug", cfg.Agent.LogLevel)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "60s", cfg.Server.Timeout)

	assert.True(t, cfg.Services.Sboxctl.Enabled)
	assert.Equal(t, []string{"sboxctl", "update", "--test"}, cfg.Services.Sboxctl.Command)
	assert.Equal(t, "15m", cfg.Services.Sboxctl.Interval)
	assert.Equal(t, "2m", cfg.Services.Sboxctl.Timeout)
	assert.True(t, cfg.Services.Sboxctl.StdoutCapture)

	assert.True(t, cfg.Services.Sboxctl.HealthCheck.Enabled)
	assert.Equal(t, "30s", cfg.Services.Sboxctl.HealthCheck.Interval)
	assert.Equal(t, "5s", cfg.Services.Sboxctl.HealthCheck.Timeout)
}

func TestLoad_WithDefaults(t *testing.T) {
	// Load config without file (should use defaults)
	cfg, err := Load("")
	require.NoError(t, err)

	// Assert default values
	assert.Equal(t, "sboxagent", cfg.Agent.Name)
	assert.Equal(t, "info", cfg.Agent.LogLevel)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "30s", cfg.Server.Timeout)

	assert.True(t, cfg.Services.Sboxctl.Enabled)
	assert.Equal(t, []string{"sboxctl", "update"}, cfg.Services.Sboxctl.Command)
	assert.Equal(t, "30m", cfg.Services.Sboxctl.Interval)
	assert.Equal(t, "5m", cfg.Services.Sboxctl.Timeout)
	assert.True(t, cfg.Services.Sboxctl.StdoutCapture)

	assert.True(t, cfg.Services.Sboxctl.HealthCheck.Enabled)
	assert.Equal(t, "1m", cfg.Services.Sboxctl.HealthCheck.Interval)
	assert.Equal(t, "10s", cfg.Services.Sboxctl.HealthCheck.Timeout)
}

func TestLoad_WithInvalidConfig(t *testing.T) {
	// Create config with missing required fields
	configContent := `
agent:
  name: ""
  version: ""
server:
  port: 99999
`

	tmpFile, err := os.CreateTemp("", "agent_invalid_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config should fail
	_, err = Load(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent name is required")
}

func TestLoad_WithInvalidPort(t *testing.T) {
	configContent := `
agent:
  name: "test"
  version: "1.0.0"
server:
  port: 99999
`

	tmpFile, err := os.CreateTemp("", "agent_invalid_port_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config should fail
	_, err = Load(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server port must be between 1 and 65535")
}

func TestLoad_WithEmptySboxctlCommand(t *testing.T) {
	configContent := `
agent:
  name: "test"
  version: "1.0.0"
services:
  sboxctl:
    enabled: true
    command: []
`

	tmpFile, err := os.CreateTemp("", "agent_empty_cmd_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config should fail
	_, err = Load(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sboxctl command is required when enabled")
}

func TestSave(t *testing.T) {
	cfg := &Config{
		Agent: AgentConfig{
			Name:     "test-save",
			Version:  "2.0.0",
			LogLevel: "warn",
		},
		Server: ServerConfig{
			Port:    9090,
			Host:    "localhost",
			Timeout: "60s",
		},
	}

	// Save config
	tmpFile := "/tmp/test_save_config.yaml"
	err := cfg.Save(tmpFile)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// Load saved config
	loadedCfg, err := Load(tmpFile)
	require.NoError(t, err)

	// Assert values match
	assert.Equal(t, cfg.Agent.Name, loadedCfg.Agent.Name)
	assert.Equal(t, cfg.Agent.Version, loadedCfg.Agent.Version)
	assert.Equal(t, "info", loadedCfg.Agent.LogLevel)
	assert.Equal(t, cfg.Server.Port, loadedCfg.Server.Port)
	assert.Equal(t, cfg.Server.Host, loadedCfg.Server.Host)
	assert.Equal(t, cfg.Server.Timeout, loadedCfg.Server.Timeout)
} 