package stamp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecute_ValidDirectories tests basic directory copying
func TestExecute_ValidDirectories(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create a regular file
	createTestFile(t, src, "readme.md", "Static content")

	// Execute
	stamper := New(nil)
	err := stamper.Execute(src, dest)

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	expectedPath := filepath.Join(dest, "readme.md")
	assertFileExists(t, expectedPath)
	assertFileContent(t, expectedPath, "Static content")
}

// TestExecute_TemplateExpansion tests .tmpl file processing
func TestExecute_TemplateExpansion(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create template file
	createTestFile(t, src, "hello.txt.tmpl", "Hello {{.name}}!")

	// Execute
	stamper := New(map[string]string{"name": "alice"})
	err := stamper.Execute(src, dest)

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	expectedPath := filepath.Join(dest, "hello.txt")
	assertFileExists(t, expectedPath)
	assertFileContent(t, expectedPath, "Hello alice!")
}

// TestExecute_TemplateExtensionRemoved tests that .tmpl extension is removed
func TestExecute_TemplateExtensionRemoved(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create template file
	createTestFile(t, src, "code.go.tmpl", "package {{.name}}")

	// Execute
	stamper := New(map[string]string{"name": "alice"})
	err := stamper.Execute(src, dest)

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// File should exist without .tmpl extension
	expectedPath := filepath.Join(dest, "code.go")
	assertFileExists(t, expectedPath)
	assertFileContent(t, expectedPath, "package alice")

	// File with .tmpl should NOT exist
	unexpectedPath := filepath.Join(dest, "code.go.tmpl")
	assertFileNotExists(t, unexpectedPath)
}

// TestExecute_NonTemplateFiles tests that regular files are copied as-is
func TestExecute_NonTemplateFiles(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create regular files
	createTestFile(t, src, "readme.md", "# README")
	createTestFile(t, src, "config.json", `{"key": "value"}`)

	// Execute
	stamper := New(nil)
	err := stamper.Execute(src, dest)

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	assertFileExists(t, filepath.Join(dest, "readme.md"))
	assertFileContent(t, filepath.Join(dest, "readme.md"), "# README")

	assertFileExists(t, filepath.Join(dest, "config.json"))
	assertFileContent(t, filepath.Join(dest, "config.json"), `{"key": "value"}`)
}

// TestExecute_NestedDirectories tests recursive directory copying
func TestExecute_NestedDirectories(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create nested directory structure
	subdir := filepath.Join(src, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	createTestFile(t, src, "root.txt", "root content")
	createTestFile(t, subdir, "nested.txt", "nested content")

	// Execute
	stamper := New(nil)
	err := stamper.Execute(src, dest)

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	assertFileExists(t, filepath.Join(dest, "root.txt"))
	assertFileContent(t, filepath.Join(dest, "root.txt"), "root content")

	assertFileExists(t, filepath.Join(dest, "subdir", "nested.txt"))
	assertFileContent(t, filepath.Join(dest, "subdir", "nested.txt"), "nested content")
}

// TestExecute_SourceNotExists tests error handling for non-existent source
func TestExecute_SourceNotExists(t *testing.T) {
	dest := t.TempDir()

	// Execute with non-existent source
	stamper := New(nil)
	err := stamper.Execute("/nonexistent/path", dest)

	// Assert error is returned
	if err == nil {
		t.Fatal("Execute() should return error for non-existent source")
	}
}

// TestExecute_InvalidTemplate tests template parsing errors
func TestExecute_InvalidTemplate(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create .tmpl file with invalid template syntax
	createTestFile(t, src, "bad.tmpl", "Invalid {{.missing")

	// Execute
	stamper := New(nil)
	err := stamper.Execute(src, dest)

	// Assert error is returned
	if err == nil {
		t.Fatal("Execute() should return error for invalid template")
	}
}

// TestExecute_MixedFiles tests both .tmpl and regular files together
func TestExecute_MixedFiles(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	// Create mix of files
	createTestFile(t, src, "greeting.txt.tmpl", "Hello {{.name}}")
	createTestFile(t, src, "readme.md", "# README")
	createTestFile(t, src, "config.tmpl", "name={{.name}}")

	// Execute
	stamper := New(map[string]string{"name": "alice"})
	err := stamper.Execute(src, dest)

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Template files should be processed
	assertFileExists(t, filepath.Join(dest, "greeting.txt"))
	assertFileContent(t, filepath.Join(dest, "greeting.txt"), "Hello alice")

	assertFileExists(t, filepath.Join(dest, "config"))
	assertFileContent(t, filepath.Join(dest, "config"), "name=alice")

	// Regular files should be copied as-is
	assertFileExists(t, filepath.Join(dest, "readme.md"))
	assertFileContent(t, filepath.Join(dest, "readme.md"), "# README")
}

// Test helpers

// createTestFile creates a file with given content in a directory
func createTestFile(t *testing.T, dir, filename, content string) string {
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

// assertFileContent verifies file exists and has expected content
func assertFileContent(t *testing.T, path, expected string) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	if string(content) != expected {
		t.Errorf("file content = %q, want %q", string(content), expected)
	}
}

// assertFileExists checks if a file exists
func assertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// assertFileNotExists checks if a file does not exist
func assertFileNotExists(t *testing.T, path string) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to not exist: %s", path)
	}
}

// TestExecute_CustomVariables tests that custom variables override defaults
func TestExecute_CustomVariables(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "hello.txt.tmpl", "Hello {{.name}}!")

	// Override default "alice" with "bob"
	customVars := map[string]string{"name": "bob"}
	stamper := New(customVars)
	err := stamper.Execute(src, dest)

	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	expectedPath := filepath.Join(dest, "hello.txt")
	assertFileContent(t, expectedPath, "Hello bob!")
}

// TestExecute_MultipleCustomVariables tests multiple custom variables
func TestExecute_MultipleCustomVariables(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "info.txt.tmpl",
		"Organization: {{.org}}, Repository: {{.repo}}")

	customVars := map[string]string{
		"org":  "monochromegane",
		"repo": "stamp",
	}
	stamper := New(customVars)
	err := stamper.Execute(src, dest)

	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "info.txt"),
		"Organization: monochromegane, Repository: stamp")
}

// TestExecute_EmptyVariables tests that empty variables result in validation error
func TestExecute_EmptyVariables(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "hello.txt.tmpl", "Hello {{.name}}!")

	// Pass empty map - should fail validation
	stamper := New(map[string]string{})
	err := stamper.Execute(src, dest)

	if err == nil {
		t.Fatal("Execute() should fail when required variables are missing")
	}

	// Should mention the missing variable
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention missing variable 'name', got: %v", err)
	}
}

// TestExecute_PartialOverride tests providing some variables but not others
func TestExecute_PartialOverride(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "mixed.txt.tmpl",
		"User: {{.name}}, Org: {{.org}}")

	// Provide both variables
	customVars := map[string]string{
		"name": "alice",
		"org":  "monochromegane",
	}
	stamper := New(customVars)
	err := stamper.Execute(src, dest)

	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "mixed.txt"),
		"User: alice, Org: monochromegane")
}

// TestExecute_StrictValidation tests that all variables must be provided
func TestExecute_StrictValidation(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "hello.tmpl", "Hello {{.name}} from {{.org}}!")

	// Only provide one of two required variables
	stamper := New(map[string]string{"name": "alice"})
	err := stamper.Execute(src, dest)

	// Should fail validation
	if err == nil {
		t.Fatal("Execute() should fail when variables are missing")
	}

	// Should be a ValidationError
	if validationErr, ok := err.(*ValidationError); !ok {
		t.Errorf("error should be ValidationError, got: %T", err)
	} else {
		// Verify the missing variable is tracked
		if _, exists := validationErr.MissingVars["org"]; !exists {
			t.Errorf("ValidationError should track missing variable 'org'")
		}
	}
}

// TestExecute_ValidationInConditionals tests variables in conditionals are required
func TestExecute_ValidationInConditionals(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "config.tmpl",
		"{{if .debug}}Debug: {{.debugLevel}}{{end}}")

	// Both variables in the if block should be required
	stamper := New(map[string]string{})
	err := stamper.Execute(src, dest)

	if err == nil {
		t.Fatal("Execute() should fail when conditional variables are missing")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "debug") {
		t.Error("should require 'debug' variable")
	}
	if !strings.Contains(errMsg, "debugLevel") {
		t.Error("should require 'debugLevel' variable")
	}
}

// TestExecute_ValidationPassesWithAllVars tests successful validation
func TestExecute_ValidationPassesWithAllVars(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "info.tmpl", "{{.name}} from {{.org}}")

	// Provide all required variables
	stamper := New(map[string]string{
		"name": "alice",
		"org":  "monochromegane",
	})
	err := stamper.Execute(src, dest)

	if err != nil {
		t.Fatalf("Execute() should succeed when all variables are provided: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "info"),
		"alice from monochromegane")
}
