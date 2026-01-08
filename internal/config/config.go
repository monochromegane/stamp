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

// LoadHierarchical loads global and template-specific configs, merging them
// Priority: template-specific > global
// Both configs are optional (returns empty map if neither exists)
func LoadHierarchical(configDir, templateName string) (map[string]string, error) {
	// Load global config (optional)
	globalPath := filepath.Join(configDir, "stamp.yaml")
	globalVars, err := loadOptional(globalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Load template-specific config (optional)
	templatePath := filepath.Join(configDir, "templates", templateName, "stamp.yaml")
	templateVars, err := loadOptional(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template config: %w", err)
	}

	// Merge configs with priority: template-specific > global
	return mergeConfigs(globalVars, templateVars), nil
}

// LoadHierarchicalMultiple loads global and multiple template-specific configs
// Priority: CLI args > rightmost template > ... > leftmost template > global
func LoadHierarchicalMultiple(configDir string, templateNames []string) (map[string]string, error) {
	// Start with global config
	globalPath := filepath.Join(configDir, "stamp.yaml")
	mergedVars, err := loadOptional(globalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Merge each template config in order (left to right)
	for _, templateName := range templateNames {
		templatePath := filepath.Join(configDir, "templates", templateName, "stamp.yaml")
		templateVars, err := loadOptional(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config for template '%s': %w", templateName, err)
		}

		// Merge with priority: current template overrides previous
		mergedVars = mergeConfigs(mergedVars, templateVars)
	}

	return mergedVars, nil
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

// mergeConfigs merges two config maps, override takes precedence
func mergeConfigs(base, override map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy base values
	for k, v := range base {
		result[k] = v
	}

	// Override with values from override map
	for k, v := range override {
		result[k] = v
	}

	return result
}
