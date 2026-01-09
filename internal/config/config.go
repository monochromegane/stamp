package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Load reads a YAML config file and returns key-value pairs
// Returns error if file doesn't exist or is invalid YAML
func Load(path string) (map[string]string, error) {
	// Check file exists first for better error message
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into map[string]string
	vars := make(map[string]string)
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return vars, nil
}

// LoadHierarchical loads global config only
// Sheet-specific configs are no longer supported
// Priority: CLI args > global config
// templateName parameter is kept for compatibility but not used
func LoadHierarchical(configDir, templateName string) (map[string]string, error) {
	return loadGlobalConfig(configDir)
}

// LoadHierarchicalMultiple loads global config for multiple sheets
// All sheets use the same global configuration
// Sheet-specific configs are no longer supported
// Priority: CLI args > global config
// templateNames parameter is kept for compatibility but not used for config loading
func LoadHierarchicalMultiple(configDir string, templateNames []string) (map[string]string, error) {
	return loadGlobalConfig(configDir)
}

// loadGlobalConfig loads the global config file from the config directory
func loadGlobalConfig(configDir string) (map[string]string, error) {
	globalPath := filepath.Join(configDir, "stamp.yaml")
	globalVars, err := loadOptional(globalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}
	return globalVars, nil
}

// loadOptional loads a config file if it exists, returns empty map if not
// Only errors on read/parse failures
func loadOptional(path string) (map[string]string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist - not an error, return empty map
		return make(map[string]string), nil
	}

	// File exists - load it using the existing Load function
	// But handle the "not found" error case (shouldn't happen given the check above)
	vars, err := Load(path)
	if err != nil {
		// If we get "not found" error here, return empty map
		// (race condition: file was deleted between Stat and Load)
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	return vars, nil
}
