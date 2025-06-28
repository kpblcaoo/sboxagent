package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
)

// ImportedConfig represents a configuration imported from sboxmgr
type ImportedConfig struct {
	Client    string                 `json:"client"`
	Version   string                 `json:"version"`
	CreatedAt string                 `json:"created_at"`
	Config    map[string]interface{} `json:"config"`
	Metadata  ConfigMetadata         `json:"metadata"`
}

// ConfigMetadata represents metadata for imported configurations
type ConfigMetadata struct {
	Source           string           `json:"source"`
	Generator        string           `json:"generator"`
	Checksum         string           `json:"checksum"`
	SubscriptionInfo SubscriptionInfo `json:"subscription_info"`
	Validation       ValidationInfo   `json:"validation,omitempty"`
}

// SubscriptionInfo represents subscription information
type SubscriptionInfo struct {
	TotalServers    int      `json:"total_servers"`
	FilteredServers int      `json:"filtered_servers"`
	ExcludedServers int      `json:"excluded_servers"`
	ExcludedList    []string `json:"excluded_list,omitempty"`
}

// ValidationInfo represents validation information
type ValidationInfo struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Importer handles importing configurations from sboxmgr
type Importer struct {
	logger *logger.Logger
	config *Config
}

// NewImporter creates a new configuration importer
func NewImporter(cfg *Config, log *logger.Logger) *Importer {
	return &Importer{
		logger: log,
		config: cfg,
	}
}

// ImportFromFile imports a configuration from a JSON file
func (i *Importer) ImportFromFile(filePath string) (*ImportedConfig, error) {
	i.logger.Info("Importing configuration from file", map[string]interface{}{
		"file": filePath,
	})

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filePath)
	}

	// Read file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer file.Close()

	return i.ImportFromReader(file)
}

// ImportFromReader imports a configuration from an io.Reader
func (i *Importer) ImportFromReader(reader io.Reader) (*ImportedConfig, error) {
	// Parse JSON
	var importedConfig ImportedConfig
	if err := json.NewDecoder(reader).Decode(&importedConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}

	// Validate imported configuration
	if err := i.validateImportedConfig(&importedConfig); err != nil {
		return nil, fmt.Errorf("invalid imported configuration: %w", err)
	}

	i.logger.Info("Configuration imported successfully", map[string]interface{}{
		"client":    importedConfig.Client,
		"version":   importedConfig.Version,
		"generator": importedConfig.Metadata.Generator,
		"servers":   importedConfig.Metadata.SubscriptionInfo.TotalServers,
	})

	return &importedConfig, nil
}

// ImportFromSboxmgr imports a configuration by executing sboxmgr
func (i *Importer) ImportFromSboxmgr(subscriptionURL, clientType string, options map[string]interface{}) (*ImportedConfig, error) {
	i.logger.Info("Importing configuration from sboxmgr", map[string]interface{}{
		"url":     subscriptionURL,
		"client":  clientType,
		"options": options,
	})

	// Build sboxmgr command
	args := []string{"json", "generate", "-u", subscriptionURL, "-c", clientType, "--no-metadata=false"}

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

	// Execute sboxmgr command
	output, err := i.executeSboxmgr(args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sboxmgr: %w", err)
	}

	// Parse output
	var importedConfig ImportedConfig
	if err := json.Unmarshal(output, &importedConfig); err != nil {
		return nil, fmt.Errorf("failed to parse sboxmgr output: %w", err)
	}

	// Validate imported configuration
	if err := i.validateImportedConfig(&importedConfig); err != nil {
		return nil, fmt.Errorf("invalid imported configuration: %w", err)
	}

	i.logger.Info("Configuration imported from sboxmgr successfully", map[string]interface{}{
		"client":    importedConfig.Client,
		"version":   importedConfig.Version,
		"generator": importedConfig.Metadata.Generator,
		"servers":   importedConfig.Metadata.SubscriptionInfo.TotalServers,
	})

	return &importedConfig, nil
}

// SaveImportedConfig saves an imported configuration to the appropriate location
func (i *Importer) SaveImportedConfig(importedConfig *ImportedConfig) error {
	// Determine target path based on client type
	targetPath, err := i.getClientConfigPath(importedConfig.Client)
	if err != nil {
		return fmt.Errorf("failed to determine config path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert to client-specific format
	clientConfig, err := i.convertToClientConfig(importedConfig)
	if err != nil {
		return fmt.Errorf("failed to convert configuration: %w", err)
	}

	// Write configuration
	if err := i.writeClientConfig(targetPath, clientConfig, importedConfig.Client); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	i.logger.Info("Configuration saved successfully", map[string]interface{}{
		"client": importedConfig.Client,
		"path":   targetPath,
	})

	return nil
}

// validateImportedConfig validates an imported configuration
func (i *Importer) validateImportedConfig(config *ImportedConfig) error {
	// Check required fields
	if config.Client == "" {
		return fmt.Errorf("client type is required")
	}

	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.Config == nil {
		return fmt.Errorf("configuration data is required")
	}

	// Validate client type
	supportedClients := []string{"sing-box", "clash", "xray", "mihomo"}
	validClient := false
	for _, client := range supportedClients {
		if config.Client == client {
			validClient = true
			break
		}
	}

	if !validClient {
		return fmt.Errorf("unsupported client type: %s", config.Client)
	}

	// Validate metadata
	if config.Metadata.Source == "" {
		return fmt.Errorf("source is required in metadata")
	}

	if config.Metadata.Generator == "" {
		return fmt.Errorf("generator is required in metadata")
	}

	return nil
}

// getClientConfigPath returns the configuration path for a client
func (i *Importer) getClientConfigPath(clientType string) (string, error) {
	switch clientType {
	case "sing-box":
		return i.config.Clients.SingBox.ConfigPath, nil
	case "xray":
		return i.config.Clients.Xray.ConfigPath, nil
	case "clash":
		return i.config.Clients.Clash.ConfigPath, nil
	case "hysteria":
		return i.config.Clients.Hysteria.ConfigPath, nil
	default:
		return "", fmt.Errorf("unsupported client type: %s", clientType)
	}
}

// convertToClientConfig converts imported config to client-specific format
func (i *Importer) convertToClientConfig(importedConfig *ImportedConfig) (interface{}, error) {
	// For now, return the config as-is
	// In the future, this could include client-specific transformations
	return importedConfig.Config, nil
}

// writeClientConfig writes configuration to file
func (i *Importer) writeClientConfig(path string, config interface{}, clientType string) error {
	// Create backup if file exists
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak." + time.Now().Format("20060102-150405")
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		i.logger.Info("Created backup of existing configuration", map[string]interface{}{
			"backup": backupPath,
		})
	}

	// Write configuration
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

// executeSboxmgr executes sboxmgr command and returns output
func (i *Importer) executeSboxmgr(args []string) ([]byte, error) {
	// This is a placeholder - in a real implementation, this would execute sboxmgr
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("sboxmgr execution not yet implemented")
}
