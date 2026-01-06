package stamp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// processTemplate reads a .tmpl file, expands it, and writes to destination
// The .tmpl extension is removed from the output filename
func (s *Stamper) processTemplate(srcPath, destPath string) error {
	// Get source file info for permissions
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat template file: %w", err)
	}

	// Remove .tmpl extension from destination
	destPath = removeTemplateExtension(destPath)

	// Read template content
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Execute template
	if err := tmpl.Execute(destFile, s.templateVars); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Set permissions to match source
	if err := os.Chmod(destPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// removeTemplateExtension strips .tmpl from the end of a path
func removeTemplateExtension(path string) string {
	if strings.HasSuffix(path, ".tmpl") {
		return strings.TrimSuffix(path, ".tmpl")
	}
	return path
}
