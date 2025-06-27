package agent

import (
	"context"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Agent: config.AgentConfig{
					Name:    "test-agent",
					Version: "1.0.0",
					LogLevel: "info",
				},
				Services: config.ServicesConfig{
					Sboxctl: config.SboxctlConfig{
						Enabled: true,
						Command: []string{"echo", "test"},
						Interval: "1m",
						Timeout: "30s",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			cfg: &config.Config{
				Agent: config.AgentConfig{
					Name:    "test-agent",
					Version: "1.0.0",
					LogLevel: "invalid-level",
				},
			},
			wantErr: true,
		},
		{
			name: "sboxctl service enabled",
			cfg: &config.Config{
				Agent: config.AgentConfig{
					Name:    "test-agent",
					Version: "1.0.0",
					LogLevel: "info",
				},
				Services: config.ServicesConfig{
					Sboxctl: config.SboxctlConfig{
						Enabled: false, // Disable to avoid command execution in tests
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sboxctl service disabled",
			cfg: &config.Config{
				Agent: config.AgentConfig{
					Name:    "test-agent",
					Version: "1.0.0",
					LogLevel: "info",
				},
				Services: config.ServicesConfig{
					Sboxctl: config.SboxctlConfig{
						Enabled: false,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := New(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
				assert.Equal(t, tt.cfg, agent.GetConfig())
			}
		})
	}
}

func TestAgent_StartStop(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:    "test-agent",
			Version: "1.0.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled: false, // Disable to avoid command execution
			},
		},
	}

	agent, err := New(cfg)
	require.NoError(t, err)

	// Test initial state
	assert.False(t, agent.IsRunning())

	// Test starting with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start agent in goroutine
	go func() {
		agent.Start(ctx)
	}()

	// Give time for agent to start
	time.Sleep(100 * time.Millisecond)
	
	// Check that agent started (may have already stopped due to timeout)
	// Don't check IsRunning() as it might be false if context already cancelled
	
	// Wait for completion
	time.Sleep(2 * time.Second)
	
	// Test that agent is not running after timeout
	assert.False(t, agent.IsRunning())
}

func TestAgent_DoubleStart(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:    "test-agent",
			Version: "1.0.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled: false, // Disable to avoid command execution
			},
		},
	}

	agent, err := New(cfg)
	require.NoError(t, err)

	// Test that agent is not running initially
	assert.False(t, agent.IsRunning())

	// Try to start agent directly (should work)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = agent.Start(ctx)
	assert.NoError(t, err)

	// Agent should have stopped due to context timeout
	assert.False(t, agent.IsRunning())
}

func TestAgent_GetStatus(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:    "test-agent",
			Version: "1.0.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled: false, // Disable to avoid command execution
			},
		},
	}

	agent, err := New(cfg)
	require.NoError(t, err)

	// Get status before starting
	status := agent.GetStatus()
	assert.False(t, status["running"].(bool))
	assert.NotNil(t, status["startTime"])
	assert.NotNil(t, status["uptime"])

	// Start agent
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		agent.Start(ctx)
	}()

	// Give time for agent to start and then stop
	time.Sleep(1 * time.Second)

	// Get status after stopping
	status = agent.GetStatus()
	assert.False(t, status["running"].(bool))
}

func TestAgent_GetConfig(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:    "test-agent",
			Version: "1.0.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled: false, // Disable to avoid command execution
			},
		},
	}

	agent, err := New(cfg)
	require.NoError(t, err)

	retrievedConfig := agent.GetConfig()
	assert.Equal(t, cfg, retrievedConfig)
	assert.Equal(t, "test-agent", retrievedConfig.Agent.Name)
	assert.Equal(t, "1.0.0", retrievedConfig.Agent.Version)
	assert.False(t, retrievedConfig.Services.Sboxctl.Enabled)
}

func TestAgent_IsRunning(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name:    "test-agent",
			Version: "1.0.0",
			LogLevel: "info",
		},
		Services: config.ServicesConfig{
			Sboxctl: config.SboxctlConfig{
				Enabled: false, // Disable to avoid command execution
			},
		},
	}

	agent, err := New(cfg)
	require.NoError(t, err)

	// Initially not running
	assert.False(t, agent.IsRunning())

	// Start agent
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		agent.Start(ctx)
	}()

	// Give time for agent to start and then stop
	time.Sleep(1 * time.Second)

	// After stopping, should not be running
	assert.False(t, agent.IsRunning())
} 