package stamp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// processTemplate reads a template file, expands it, and writes to destination
// The template extension is removed from the output filename
func (s *Stamper) processTemplate(srcPath, destPath string) error {
	// Remove custom extension from destination
	destPath = s.removeTemplateExtension(destPath)

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

	return nil
}

// removeTemplateExtension strips the template extension from the end of a path
func (s *Stamper) removeTemplateExtension(path string) string {
	if strings.HasSuffix(path, s.templateExt) {
		return strings.TrimSuffix(path, s.templateExt)
	}
	return path
}
