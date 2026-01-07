package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	if cli == nil {
		t.Error("NewCLI() returned nil")
	}
}

func TestPressCmd_WithTemplateConfig(t *testing.T) {
	// Setup config directory structure
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory structure
	templateDir := filepath.Join(configDir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template file
	tmplPath := filepath.Join(templateDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create template-specific config
	configPath := filepath.Join(templateDir, "stamp.yaml")
	configContent := `name: charlie`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Execute CLI
	cli := NewCLI()
	args := []string{"-t", "go-cli", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	resultPath := filepath.Join(destDir, "hello.txt")
	content, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	expected := "Hello charlie!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q", string(content), expected)
	}
}

func TestPressCmd_HierarchicalConfig(t *testing.T) {
	// Setup config directory structure
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template file
	tmplPath := filepath.Join(templateDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}} from {{.org}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create global config
	globalConfigPath := filepath.Join(configDir, "stamp.yaml")
	globalConfig := `org: global-org
author: alice`
	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("failed to create global config: %v", err)
	}

	// Create template-specific config (overrides org)
	templateConfigPath := filepath.Join(templateDir, "stamp.yaml")
	templateConfig := `name: bob
org: template-org`
	if err := os.WriteFile(templateConfigPath, []byte(templateConfig), 0644); err != nil {
		t.Fatalf("failed to create template config: %v", err)
	}

	// Execute CLI
	cli := NewCLI()
	args := []string{"-t", "go-cli", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	// Template config should override global (org=template-org, not global-org)
	expected := "Hello bob from template-org!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q (template config should override global)", string(content), expected)
	}
}

func TestPressCmd_CLIArgsOverrideConfig(t *testing.T) {
	// Setup config directory structure
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template
	tmplPath := filepath.Join(templateDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create config with name=charlie
	configPath := filepath.Join(templateDir, "stamp.yaml")
	if err := os.WriteFile(configPath, []byte("name: charlie\n"), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Execute with CLI override: name=dave
	cli := NewCLI()
	args := []string{"-t", "go-cli", "-d", destDir, "-c", configDir, "name=dave"}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	// CLI arg should win
	expected := "Hello dave!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q (CLI should override config)", string(content), expected)
	}
}

func TestPressCmd_WithoutConfig(t *testing.T) {
	// Setup config directory structure (no config files)
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template
	tmplPath := filepath.Join(templateDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Execute with CLI variables only (no config files)
	cli := NewCLI()
	args := []string{"-t", "go-cli", "-d", destDir, "-c", configDir, "name=frank"}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	expected := "Hello frank!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q", string(content), expected)
	}
}

func TestPressCmd_InvalidConfigDirectory(t *testing.T) {
	destDir := t.TempDir()

	// Execute with non-existent config directory
	cli := NewCLI()
	args := []string{"-t", "go-cli", "-d", destDir, "-c", "/nonexistent/config"}
	err := cli.Execute(args)

	// Should fail
	if err == nil {
		t.Fatal("Execute() should fail with non-existent config directory")
	}
	if !strings.Contains(err.Error(), "config directory not found") {
		t.Errorf("error should mention config directory not found, got: %v", err)
	}
}

func TestPressCmd_InvalidTemplateName(t *testing.T) {
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create templates directory but no templates
	templatesDir := filepath.Join(configDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	// Execute with non-existent template
	cli := NewCLI()
	args := []string{"-t", "does-not-exist", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Should fail
	if err == nil {
		t.Fatal("Execute() should fail with non-existent template")
	}
	if !strings.Contains(err.Error(), "template 'does-not-exist' not found") {
		t.Errorf("error should mention template not found, got: %v", err)
	}
}

func TestPressCmd_MissingVariables(t *testing.T) {
	// Setup config directory structure (no config files)
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template with required variable
	tmplPath := filepath.Join(templateDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Execute without providing required variables
	cli := NewCLI()
	args := []string{"-t", "go-cli", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Assert - should fail with strict validation
	if err == nil {
		t.Fatal("Execute() should fail when required variables are missing")
	}

	// Error should mention the missing variable
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention missing variable 'name', got: %v", err)
	}
	if !strings.Contains(err.Error(), "missing required template variables") {
		t.Errorf("error should indicate missing variables, got: %v", err)
	}
}
