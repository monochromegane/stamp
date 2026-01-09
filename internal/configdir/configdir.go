package configdir

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GetConfigDir returns the default config directory path
// Priority: $XDG_CONFIG_HOME/stamp > os.UserConfigDir()/stamp
// Does NOT create the directory
func GetConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "stamp"), nil
	}

	// Fall back to os.UserConfigDir()
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user config directory: %w", err)
	}

	return filepath.Join(userConfig, "stamp"), nil
}

// GetConfigDirWithOverride returns config directory, with optional override
// If override is empty, uses GetConfigDir()
// If override is provided, validates it exists and returns it
func GetConfigDirWithOverride(override string) (string, error) {
	if override == "" {
		return GetConfigDir()
	}

	// Validate override path exists
	info, err := os.Stat(override)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("config directory not found: %s", override)
	}
	if err != nil {
		return "", fmt.Errorf("failed to access config directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("config path is not a directory: %s", override)
	}

	return override, nil
}

// ResolveTemplateDir resolves sheet directory path and validates existence
// Returns: {configDir}/sheets/{templateName}/
// Validates directory exists, returns helpful error if not
func ResolveTemplateDir(configDir, templateName string) (string, error) {
	templatePath := filepath.Join(configDir, "sheets", templateName)

	// Check if sheet directory exists
	info, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		// Sheet doesn't exist - provide helpful error with available sheets
		available, listErr := ListAvailableSheets(configDir)
		if listErr != nil || len(available) == 0 {
			return "", fmt.Errorf("sheet '%s' not found in %s/sheets/\n\nCreate sheet directory: mkdir -p %s/sheets/%s",
				templateName, configDir, configDir, templateName)
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("sheet '%s' not found in %s/sheets/\n\n", templateName, configDir))
		sb.WriteString("Available sheets:\n")
		for _, name := range available {
			sb.WriteString(fmt.Sprintf("  - %s\n", name))
		}
		sb.WriteString(fmt.Sprintf("\nCreate new sheet: mkdir -p %s/sheets/%s", configDir, templateName))
		return "", fmt.Errorf("%s", sb.String())
	}
	if err != nil {
		return "", fmt.Errorf("failed to access sheet directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("sheet path is not a directory: %s", templatePath)
	}

	return templatePath, nil
}

// ListAvailableSheets returns list of sheet names in config directory
// Returns: []string of sheet names from sheets/ subdirectory
// Used for error messages when sheet not found
func ListAvailableSheets(configDir string) ([]string, error) {
	sheetsDir := filepath.Join(configDir, "sheets")

	// Check if sheets directory exists
	info, err := os.Stat(sheetsDir)
	if os.IsNotExist(err) {
		// Sheets directory doesn't exist - return empty slice (not an error)
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access sheets directory: %w", err)
	}
	if !info.IsDir() {
		return []string{}, nil
	}

	// Read sheets directory
	entries, err := os.ReadDir(sheetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheets directory: %w", err)
	}

	// Collect directory names (sheets)
	var sheets []string
	for _, entry := range entries {
		if entry.IsDir() {
			sheets = append(sheets, entry.Name())
		}
	}

	// Sort for consistent output
	sort.Strings(sheets)
	return sheets, nil
}

// ResolveTemplateDirs resolves multiple sheet directories and validates ALL exist
// Returns all resolved paths OR comprehensive error
func ResolveTemplateDirs(configDir string, templateNames []string) ([]string, error) {
	if len(templateNames) == 0 {
		return nil, fmt.Errorf("no sheets specified")
	}

	var resolvedPaths []string
	var missingTemplates []string
	var foundTemplates []string

	// Try to resolve each sheet
	for _, name := range templateNames {
		path := filepath.Join(configDir, "sheets", name)
		info, err := os.Stat(path)

		if os.IsNotExist(err) {
			missingTemplates = append(missingTemplates, name)
			foundTemplates = append(foundTemplates, fmt.Sprintf("  ✗ %s - not found", name))
		} else if err != nil {
			return nil, fmt.Errorf("failed to access sheet '%s': %w", name, err)
		} else if !info.IsDir() {
			return nil, fmt.Errorf("sheet path is not a directory: %s", path)
		} else {
			resolvedPaths = append(resolvedPaths, path)
			foundTemplates = append(foundTemplates, fmt.Sprintf("  ✓ %s - %s", name, path))
		}
	}

	// If any sheets are missing, return comprehensive error
	if len(missingTemplates) > 0 {
		available, _ := ListAvailableSheets(configDir)

		var sb strings.Builder
		sb.WriteString("Failed to resolve sheets:\n")
		for _, line := range foundTemplates {
			sb.WriteString(line + "\n")
		}

		if len(available) > 0 {
			sb.WriteString("\nAvailable sheets:\n")
			for _, name := range available {
				sb.WriteString(fmt.Sprintf("  - %s\n", name))
			}
		}

		sb.WriteString("\nCreate missing sheets:\n")
		for _, name := range missingTemplates {
			sb.WriteString(fmt.Sprintf("  mkdir -p %s/sheets/%s\n", configDir, name))
		}

		return nil, fmt.Errorf("%s", sb.String())
	}

	return resolvedPaths, nil
}
