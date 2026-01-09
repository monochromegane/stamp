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
	templateExt  string // Stamp file extension (e.g., ".stamp", ".tmpl", ".tpl")
}

// New creates a new Stamper with provided template variables and extension
func New(vars map[string]string, ext string) *Stamper {
	templateVars := make(map[string]string)
	for k, v := range vars {
		templateVars[k] = v
	}

	// Default to .stamp if not specified
	if ext == "" {
		ext = ".stamp"
	}

	return &Stamper{
		templateVars: templateVars,
		templateExt:  ext,
	}
}

// Execute performs the directory copy operation
// It walks the source directory tree and processes each file
func (s *Stamper) Execute(src, dest string) error {
	return s.ExecuteMultiple([]string{src}, dest)
}

// ExecuteMultiple processes multiple template directories sequentially
// Later templates overwrite files from earlier templates
func (s *Stamper) ExecuteMultiple(srcDirs []string, dest string) error {
	if len(srcDirs) == 0 {
		return fmt.Errorf("no source directories provided")
	}

	// Pre-validate ALL template variables across all templates
	if err := s.validateMultipleTemplateVars(srcDirs); err != nil {
		return err
	}

	// Create destination directory once
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Process each template sequentially
	for i, src := range srcDirs {
		// Validate source exists
		srcInfo, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("source directory error (template %d): %w", i+1, err)
		}
		if !srcInfo.IsDir() {
			return fmt.Errorf("source is not a directory (template %d): %s", i+1, src)
		}

		// Walk and process this template directory
		if err := s.processTemplateDir(src, dest); err != nil {
			return fmt.Errorf("failed to process template %d (%s): %w", i+1, src, err)
		}
	}

	return nil
}

// processTemplateDir walks a single template directory and processes files
func (s *Stamper) processTemplateDir(src, dest string) error {
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

// isTmplNoopFile checks if a file ends with the template extension plus .noop
func (s *Stamper) isTmplNoopFile(path string) bool {
	return strings.HasSuffix(path, s.templateExt+".noop")
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
	// Check .{ext}.noop first (more specific)
	if s.isTmplNoopFile(srcPath) {
		return s.processTmplNoop(srcPath, destPath)
	}

	// Check if file ends with custom extension
	if strings.HasSuffix(srcPath, s.templateExt) {
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
