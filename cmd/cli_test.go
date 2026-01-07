package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	if cli == nil {
		t.Error("NewCLI() returned nil")
	}
}

func TestPressCmd_WithConfigFile(t *testing.T) {
	// Setup
	srcDir := t.TempDir()
	destDir := t.TempDir()
	configDir := t.TempDir()

	// Create template file
	tmplPath := filepath.Join(srcDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create config file
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `name: charlie`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Execute CLI
	cli := NewCLI()
	args := []string{"press", "-s", srcDir, "-d", destDir, "-c", configPath}
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

func TestPressCmd_CLIArgsOverrideConfig(t *testing.T) {
	// Setup
	srcDir := t.TempDir()
	destDir := t.TempDir()
	configDir := t.TempDir()

	// Create template
	tmplPath := filepath.Join(srcDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create config with name=charlie
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("name: charlie\n"), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Execute with CLI override: name=dave
	cli := NewCLI()
	args := []string{"press", "-s", srcDir, "-d", destDir, "-c", configPath, "name=dave"}
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

func TestPressCmd_ConfigOverridesDefaults(t *testing.T) {
	// Setup
	srcDir := t.TempDir()
	destDir := t.TempDir()
	configDir := t.TempDir()

	// Create template
	tmplPath := filepath.Join(srcDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create config
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("name: eve\n"), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Execute with config only (no CLI args)
	cli := NewCLI()
	args := []string{"press", "-s", srcDir, "-d", destDir, "-c", configPath}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	// Config should override default "alice"
	expected := "Hello eve!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q (config should override default)", string(content), expected)
	}
}

func TestPressCmd_InvalidConfigFile(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Execute with non-existent config
	cli := NewCLI()
	args := []string{"press", "-s", srcDir, "-d", destDir, "-c", "/nonexistent/config.yaml"}
	err := cli.Execute(args)

	// Should fail
	if err == nil {
		t.Fatal("Execute() should fail with non-existent config file")
	}
}

func TestPressCmd_WithoutConfig_BackwardCompatible(t *testing.T) {
	// Setup
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create template
	tmplPath := filepath.Join(srcDir, "hello.txt.tmpl")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Execute without config and without variables
	cli := NewCLI()
	args := []string{"press", "-s", srcDir, "-d", destDir}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	// Without config or CLI args, templates show <no value>
	expected := "Hello <no value>!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q (no variables provided)", string(content), expected)
	}
}
