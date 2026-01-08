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

// ResolveTemplateDir resolves template directory path and validates existence
// Returns: {configDir}/templates/{templateName}/
// Validates directory exists, returns helpful error if not
func ResolveTemplateDir(configDir, templateName string) (string, error) {
	templatePath := filepath.Join(configDir, "templates", templateName)

	// Check if template directory exists
	info, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		// Template doesn't exist - provide helpful error with available templates
		available, listErr := ListAvailableTemplates(configDir)
		if listErr != nil || len(available) == 0 {
			return "", fmt.Errorf("template '%s' not found in %s/templates/\n\nCreate template directory: mkdir -p %s/templates/%s",
				templateName, configDir, configDir, templateName)
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("template '%s' not found in %s/templates/\n\n", templateName, configDir))
		sb.WriteString("Available templates:\n")
		for _, name := range available {
			sb.WriteString(fmt.Sprintf("  - %s\n", name))
		}
		sb.WriteString(fmt.Sprintf("\nCreate new template: mkdir -p %s/templates/%s", configDir, templateName))
		return "", fmt.Errorf("%s", sb.String())
	}
	if err != nil {
		return "", fmt.Errorf("failed to access template directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("template path is not a directory: %s", templatePath)
	}

	return templatePath, nil
}

// ListAvailableTemplates returns list of template names in config directory
// Returns: []string of template names from templates/ subdirectory
// Used for error messages when template not found
func ListAvailableTemplates(configDir string) ([]string, error) {
	templatesDir := filepath.Join(configDir, "templates")

	// Check if templates directory exists
	info, err := os.Stat(templatesDir)
	if os.IsNotExist(err) {
		// Templates directory doesn't exist - return empty slice (not an error)
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access templates directory: %w", err)
	}
	if !info.IsDir() {
		return []string{}, nil
	}

	// Read templates directory
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	// Collect directory names (templates)
	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}

	// Sort for consistent output
	sort.Strings(templates)
	return templates, nil
}

// ResolveTemplateDirs resolves multiple template directories and validates ALL exist
// Returns all resolved paths OR comprehensive error
func ResolveTemplateDirs(configDir string, templateNames []string) ([]string, error) {
	if len(templateNames) == 0 {
		return nil, fmt.Errorf("no templates specified")
	}

	var resolvedPaths []string
	var missingTemplates []string
	var foundTemplates []string

	// Try to resolve each template
	for _, name := range templateNames {
		path := filepath.Join(configDir, "templates", name)
		info, err := os.Stat(path)

		if os.IsNotExist(err) {
			missingTemplates = append(missingTemplates, name)
			foundTemplates = append(foundTemplates, fmt.Sprintf("  ✗ %s - not found", name))
		} else if err != nil {
			return nil, fmt.Errorf("failed to access template '%s': %w", name, err)
		} else if !info.IsDir() {
			return nil, fmt.Errorf("template path is not a directory: %s", path)
		} else {
			resolvedPaths = append(resolvedPaths, path)
			foundTemplates = append(foundTemplates, fmt.Sprintf("  ✓ %s - %s", name, path))
		}
	}

	// If any templates are missing, return comprehensive error
	if len(missingTemplates) > 0 {
		available, _ := ListAvailableTemplates(configDir)

		var sb strings.Builder
		sb.WriteString("Failed to resolve templates:\n")
		for _, line := range foundTemplates {
			sb.WriteString(line + "\n")
		}

		if len(available) > 0 {
			sb.WriteString("\nAvailable templates:\n")
			for _, name := range available {
				sb.WriteString(fmt.Sprintf("  - %s\n", name))
			}
		}

		sb.WriteString("\nCreate missing templates:\n")
		for _, name := range missingTemplates {
			sb.WriteString(fmt.Sprintf("  mkdir -p %s/templates/%s\n", configDir, name))
		}

		return nil, fmt.Errorf("%s", sb.String())
	}

	return resolvedPaths, nil
}
