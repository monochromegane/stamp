package stamp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Stamper handles directory copying with template expansion
type Stamper struct {
	templateVars map[string]string
}

// New creates a new Stamper with provided template variables
func New(vars map[string]string) *Stamper {
	templateVars := make(map[string]string)
	for k, v := range vars {
		templateVars[k] = v
	}

	return &Stamper{
		templateVars: templateVars,
	}
}

// Execute performs the directory copy operation
// It walks the source directory tree and processes each file
func (s *Stamper) Execute(src, dest string) error {
	// Validate source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source directory error: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Validate template variables before any processing
	if err := s.validateTemplateVars(src); err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from source
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Calculate destination path
		destPath := filepath.Join(dest, relPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Handle files
		return s.processFile(path, destPath)
	})
}

// isTmplNoopFile checks if a file ends with .tmpl.noop
func isTmplNoopFile(path string) bool {
	return strings.HasSuffix(path, ".tmpl.noop")
}

// removeNoopExtension strips .noop from the end of a path
func removeNoopExtension(path string) string {
	if strings.HasSuffix(path, ".noop") {
		return strings.TrimSuffix(path, ".noop")
	}
	return path
}

// processFile determines whether to template or copy a file
func (s *Stamper) processFile(srcPath, destPath string) error {
	// Check .tmpl.noop first (more specific)
	if isTmplNoopFile(srcPath) {
		return s.processTmplNoop(srcPath, destPath)
	}

	// Check if file ends with .tmpl
	if strings.HasSuffix(srcPath, ".tmpl") {
		return s.processTemplate(srcPath, destPath)
	}
	return s.copyFile(srcPath, destPath)
}

// copyFile copies a regular file from src to dest
func (s *Stamper) copyFile(src, dest string) error {
	// Read source file
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination with standard permissions
	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// processTmplNoop copies a .tmpl.noop file, removing only the .noop extension
// This allows template files to be included in output without variable expansion
func (s *Stamper) processTmplNoop(srcPath, destPath string) error {
	// Remove .noop extension from destination (keeping .tmpl)
	destPath = removeNoopExtension(destPath)

	// Copy file as-is without template processing
	return s.copyFile(srcPath, destPath)
}
