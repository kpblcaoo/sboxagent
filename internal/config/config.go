package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the main configuration structure
type Config struct {
	Agent    AgentConfig    `mapstructure:"agent"`
	Server   ServerConfig   `mapstructure:"server"`
	Services ServicesConfig `mapstructure:"services"`
	Clients  ClientsConfig  `mapstructure:"clients"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Security SecurityConfig `mapstructure:"security"`
}

// AgentConfig represents agent basic configuration
type AgentConfig struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	LogLevel string `mapstructure:"log_level"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
	Timeout string `mapstructure:"timeout"`
}

// ServicesConfig represents service management configuration
type ServicesConfig struct {
	Sboxctl SboxctlConfig `mapstructure:"sboxctl"`
}

// SboxctlConfig represents sboxctl service configuration
type SboxctlConfig struct {
	Enabled       bool              `mapstructure:"enabled"`
	Command       []string          `mapstructure:"command"`
	Interval      string            `mapstructure:"interval"`
	Timeout       string            `mapstructure:"timeout"`
	StdoutCapture bool              `mapstructure:"stdout_capture"`
	HealthCheck   HealthCheckConfig `mapstructure:"health_check"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Interval string `mapstructure:"interval"`
	Timeout  string `mapstructure:"timeout"`
}

// ClientsConfig represents VPN client configuration
type ClientsConfig struct {
	SingBox SingBoxConfig `mapstructure:"sing-box"`
	Xray    XrayConfig    `mapstructure:"xray"`
	Clash   ClashConfig   `mapstructure:"clash"`
	Hysteria HysteriaConfig `mapstructure:"hysteria"`
}

// SingBoxConfig represents sing-box client configuration
type SingBoxConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	BinaryPath string `mapstructure:"binary_path"`
	ConfigPath string `mapstructure:"config_path"`
}

// XrayConfig represents xray client configuration
type XrayConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	BinaryPath string `mapstructure:"binary_path"`
	ConfigPath string `mapstructure:"config_path"`
}

// ClashConfig represents clash client configuration
type ClashConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	BinaryPath string `mapstructure:"binary_path"`
	ConfigPath string `mapstructure:"config_path"`
}

// HysteriaConfig represents hysteria client configuration
type HysteriaConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	BinaryPath string `mapstructure:"binary_path"`
	ConfigPath string `mapstructure:"config_path"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	StdoutCapture bool `mapstructure:"stdout_capture"`
	Aggregation   bool `mapstructure:"aggregation"`
	RetentionDays int  `mapstructure:"retention_days"`
	MaxEntries    int  `mapstructure:"max_entries"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	AllowRemoteAPI bool     `mapstructure:"allow_remote_api"`
	APIToken       string   `mapstructure:"api_token"`
	AllowedHosts   []string `mapstructure:"allowed_hosts"`
	TLSEnabled     bool     `mapstructure:"tls_enabled"`
	TLSCertFile    string   `mapstructure:"tls_cert_file"`
	TLSKeyFile     string   `mapstructure:"tls_key_file"`
}

// Load loads configuration from file or creates default
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// If config file is provided, read it
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Try to find config in common locations
		v.SetConfigName("agent")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/sboxagent")
		v.AddConfigPath("$HOME/.sboxagent")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
			// Config file not found, use defaults
			fmt.Println("No configuration file found, using defaults")
		}
	}

	// Environment variable overrides
	v.SetEnvPrefix("SBOXAGENT")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Agent defaults
	v.SetDefault("agent.name", "sboxagent")
	v.SetDefault("agent.log_level", "info")

	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.timeout", "30s")

	// Services defaults
	v.SetDefault("services.sboxctl.enabled", true)
	v.SetDefault("services.sboxctl.command", []string{"sboxctl", "update"})
	v.SetDefault("services.sboxctl.interval", "30m")
	v.SetDefault("services.sboxctl.timeout", "5m")
	v.SetDefault("services.sboxctl.stdout_capture", true)
	v.SetDefault("services.sboxctl.health_check.enabled", true)
	v.SetDefault("services.sboxctl.health_check.interval", "1m")
	v.SetDefault("services.sboxctl.health_check.timeout", "10s")

	// Clients defaults
	v.SetDefault("clients.sing-box.enabled", true)
	v.SetDefault("clients.sing-box.binary_path", "/usr/local/bin/sing-box")
	v.SetDefault("clients.sing-box.config_path", "/etc/sing-box/config.json")

	v.SetDefault("clients.xray.enabled", true)
	v.SetDefault("clients.xray.binary_path", "/usr/local/bin/xray")
	v.SetDefault("clients.xray.config_path", "/etc/xray/config.json")

	v.SetDefault("clients.clash.enabled", true)
	v.SetDefault("clients.clash.binary_path", "/usr/local/bin/clash")
	v.SetDefault("clients.clash.config_path", "/etc/clash/config.yaml")

	v.SetDefault("clients.hysteria.enabled", true)
	v.SetDefault("clients.hysteria.binary_path", "/usr/local/bin/hysteria")
	v.SetDefault("clients.hysteria.config_path", "/etc/hysteria/config.json")

	// Logging defaults
	v.SetDefault("logging.stdout_capture", true)
	v.SetDefault("logging.aggregation", true)
	v.SetDefault("logging.retention_days", 30)
	v.SetDefault("logging.max_entries", 1000)

	// Security defaults
	v.SetDefault("security.allow_remote_api", false)
	v.SetDefault("security.allowed_hosts", []string{"127.0.0.1", "::1"})
	v.SetDefault("security.tls_enabled", false)
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	// Validate agent configuration
	if cfg.Agent.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	if cfg.Agent.Version == "" {
		return fmt.Errorf("agent version is required")
	}

	// Validate server configuration
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	// Validate sboxctl configuration if enabled
	if cfg.Services.Sboxctl.Enabled {
		if len(cfg.Services.Sboxctl.Command) == 0 {
			return fmt.Errorf("sboxctl command is required when enabled")
		}
	}

	return nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	v := viper.New()
	
	// Convert config back to map
	if err := v.MergeConfigMap(map[string]interface{}{
		"agent":    c.Agent,
		"server":   c.Server,
		"services": c.Services,
		"clients":  c.Clients,
		"logging":  c.Logging,
		"security": c.Security,
	}); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	// Ensure directory exists
	dir := path[:len(path)-len("/"+path[len(path)-1:])]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return v.WriteConfigAs(path)
} 