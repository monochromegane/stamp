package config

import (
	"fmt"
	"os"

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
