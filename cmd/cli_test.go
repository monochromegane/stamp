package cmd

import (
	"bytes"
	"io"
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

func TestPressCmd_CLIArgsOverrideGlobalConfig(t *testing.T) {
	// Setup config directory structure
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "sheets", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template
	tmplPath := filepath.Join(templateDir, "hello.txt.stamp")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create GLOBAL config with name=charlie (not sheet-specific)
	globalConfigPath := filepath.Join(configDir, "stamp.yaml")
	if err := os.WriteFile(globalConfigPath, []byte("name: charlie\n"), 0644); err != nil {
		t.Fatalf("failed to create global config: %v", err)
	}

	// Execute with CLI override: name=dave
	cli := NewCLI()
	args := []string{"-s", "go-cli", "-d", destDir, "-c", configDir, "name=dave"}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	// CLI arg should win over global config
	expected := "Hello dave!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q (CLI should override global config)", string(content), expected)
	}
}

func TestPressCmd_GlobalConfigOnly(t *testing.T) {
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "sheets", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template file
	tmplPath := filepath.Join(templateDir, "hello.txt.stamp")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}} from {{.org}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Create global config
	globalConfigPath := filepath.Join(configDir, "stamp.yaml")
	globalConfig := `name: alice
org: global-org`
	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("failed to create global config: %v", err)
	}

	// Create sheet-specific config (should be ignored)
	sheetConfigPath := filepath.Join(templateDir, "stamp.yaml")
	sheetConfig := `name: should-be-ignored
org: should-be-ignored`
	if err := os.WriteFile(sheetConfigPath, []byte(sheetConfig), 0644); err != nil {
		t.Fatalf("failed to create sheet config: %v", err)
	}

	// Execute CLI
	cli := NewCLI()
	args := []string{"-s", "go-cli", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	// Should use global config values, not sheet config
	expected := "Hello alice from global-org!"
	if string(content) != expected {
		t.Errorf("content = %q, want %q (should use global config, not sheet config)", string(content), expected)
	}
}

func TestPressCmd_MultipleSheets_GlobalConfigOnly(t *testing.T) {
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create global config
	globalConfigPath := filepath.Join(configDir, "stamp.yaml")
	globalConfig := `name: alice
org: global-org
version: 1.0.0`
	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644); err != nil {
		t.Fatalf("failed to create global config: %v", err)
	}

	// Create multiple sheets with their own configs (should all be ignored)
	sheets := []string{"base", "backend"}
	for _, sheetName := range sheets {
		sheetDir := filepath.Join(configDir, "sheets", sheetName)
		if err := os.MkdirAll(sheetDir, 0755); err != nil {
			t.Fatalf("failed to create sheet dir: %v", err)
		}

		// Create sheet config (should be ignored)
		sheetConfigPath := filepath.Join(sheetDir, "stamp.yaml")
		sheetConfig := "name: " + sheetName + "-name\norg: " + sheetName + "-org"
		if err := os.WriteFile(sheetConfigPath, []byte(sheetConfig), 0644); err != nil {
			t.Fatalf("failed to create sheet config: %v", err)
		}

		// Create template
		tmplPath := filepath.Join(sheetDir, sheetName+".txt.stamp")
		if err := os.WriteFile(tmplPath, []byte("{{.name}} from {{.org}} v{{.version}}"), 0644); err != nil {
			t.Fatalf("failed to create template: %v", err)
		}
	}

	// Execute with multiple sheets
	cli := NewCLI()
	args := []string{"-s", "base", "-s", "backend", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Check both templates used global config
	for _, sheetName := range sheets {
		filePath := filepath.Join(destDir, sheetName+".txt")
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", filePath, err)
		}

		expected := "alice from global-org v1.0.0"
		if string(content) != expected {
			t.Errorf("%s content = %q, want %q (should use global config)", sheetName, string(content), expected)
		}
	}
}

func TestPressCmd_WithoutConfig(t *testing.T) {
	// Setup config directory structure (no config files)
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "sheets", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template
	tmplPath := filepath.Join(templateDir, "hello.txt.stamp")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Execute with CLI variables only (no config files)
	cli := NewCLI()
	args := []string{"-s", "go-cli", "-d", destDir, "-c", configDir, "name=frank"}
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
	args := []string{"-s", "go-cli", "-d", destDir, "-c", "/nonexistent/config"}
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

	// Create sheets directory but no sheets
	sheetsDir := filepath.Join(configDir, "sheets")
	if err := os.MkdirAll(sheetsDir, 0755); err != nil {
		t.Fatalf("failed to create sheets dir: %v", err)
	}

	// Execute with non-existent template
	cli := NewCLI()
	args := []string{"-s", "does-not-exist", "-d", destDir, "-c", configDir}
	err := cli.Execute(args)

	// Should fail
	if err == nil {
		t.Fatal("Execute() should fail with non-existent template")
	}
	if !strings.Contains(err.Error(), "does-not-exist") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention template not found, got: %v", err)
	}
}

func TestPressCmd_MissingVariables(t *testing.T) {
	// Setup config directory structure (no config files)
	configDir := t.TempDir()
	destDir := t.TempDir()

	// Create template directory
	templateDir := filepath.Join(configDir, "sheets", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create template with required variable
	tmplPath := filepath.Join(templateDir, "hello.txt.stamp")
	if err := os.WriteFile(tmplPath, []byte("Hello {{.name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	// Execute without providing required variables
	cli := NewCLI()
	args := []string{"-s", "go-cli", "-d", destDir, "-c", configDir}
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

func TestConfigDirCmd_DefaultPath(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cli := NewCLI()
	err := cli.Execute([]string{"config-dir"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output is a valid path ending with stamp
	if filepath.Base(strings.TrimSpace(output)) != "stamp" {
		t.Errorf("output = %q, want path ending with stamp", output)
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(strings.TrimSpace(output)) {
		t.Errorf("output = %q, want absolute path", output)
	}
}

func TestConfigDirCmd_OverridePath(t *testing.T) {
	configDir := t.TempDir()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cli := NewCLI()
	err := cli.Execute([]string{"config-dir", "-c", configDir})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	if output != configDir {
		t.Errorf("output = %q, want %q", output, configDir)
	}
}

func TestConfigDirCmd_InvalidPath(t *testing.T) {
	cli := NewCLI()
	err := cli.Execute([]string{"config-dir", "-c", "/nonexistent/path"})

	if err == nil {
		t.Fatal("Execute() succeeded, want error")
	}

	if !strings.Contains(err.Error(), "config directory not found") {
		t.Errorf("error = %q, want error containing 'config directory not found'", err.Error())
	}
}

func TestCollectCmd_BasicDirectory(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source files
	file1 := filepath.Join(sourceDir, "file1.txt")
	file2 := filepath.Join(sourceDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// Execute collect
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify files were copied
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt not found: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sheetDir, "file2.txt")); err != nil {
		t.Errorf("file2.txt not found: %v", err)
	}
}

func TestCollectCmd_SkipGitDirectory(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source files and .git directory
	file1 := filepath.Join(sourceDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}

	gitDir := filepath.Join(sourceDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	gitConfig := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfig, []byte("git-config"), 0644); err != nil {
		t.Fatalf("failed to create .git/config: %v", err)
	}

	// Execute collect
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify .git directory was skipped
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should not be copied")
	}

	// Verify regular file was copied
	if _, err := os.Stat(filepath.Join(sheetDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt not found: %v", err)
	}
}

func TestCollectCmd_SkipGitFile(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source files and .git file (git worktree)
	file1 := filepath.Join(sourceDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}

	gitFile := filepath.Join(sourceDir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: /path/to/worktree"), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	// Execute collect
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify .git file was skipped
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git file should not be copied")
	}

	// Verify regular file was copied
	if _, err := os.Stat(filepath.Join(sheetDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt not found: %v", err)
	}
}

func TestCollectCmd_TemplateFlag(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source file
	file1 := filepath.Join(sourceDir, "template.txt")
	if err := os.WriteFile(file1, []byte("{{.name}}"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Execute collect with template flag
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-t", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify file has .stamp extension
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, "template.txt.stamp")); err != nil {
		t.Errorf("template.txt.stamp not found: %v", err)
	}

	// Verify original name doesn't exist
	if _, err := os.Stat(filepath.Join(sheetDir, "template.txt")); !os.IsNotExist(err) {
		t.Error("template.txt should not exist (should be .stamp)")
	}
}

func TestCollectCmd_CustomExtension(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source file
	file1 := filepath.Join(sourceDir, "template.txt")
	if err := os.WriteFile(file1, []byte("{{.name}}"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Execute collect with custom extension
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-t", "-e", ".tmpl", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify file has custom extension
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, "template.txt.tmpl")); err != nil {
		t.Errorf("template.txt.tmpl not found: %v", err)
	}
}

func TestCollectCmd_SheetAlreadyExists(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create existing sheet
	sheetDir := filepath.Join(configDir, "sheets", "existing-sheet")
	if err := os.MkdirAll(sheetDir, 0755); err != nil {
		t.Fatalf("failed to create existing sheet: %v", err)
	}

	// Create source file
	file1 := filepath.Join(sourceDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Execute collect with existing sheet name
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "existing-sheet", "-c", configDir, sourceDir})

	// Assert - should fail
	if err == nil {
		t.Fatal("Execute() should fail with existing sheet")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestCollectCmd_NonExistentSource(t *testing.T) {
	// Setup config directory
	configDir := t.TempDir()

	// Execute collect with non-existent source
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir, "/nonexistent/source"})

	// Assert - should fail
	if err == nil {
		t.Fatal("Execute() should fail with non-existent source")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestCollectCmd_DefaultSourceCurrentDir(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source file
	file1 := filepath.Join(sourceDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Change to source directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(sourceDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute collect without source argument (should use current directory)
	cli := NewCLI()
	err = cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify file was copied
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt not found: %v", err)
	}
}

func TestCollectCmd_NestedDirectories(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(sourceDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// Execute collect
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify nested structure was preserved
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt not found: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sheetDir, "subdir", "file2.txt")); err != nil {
		t.Errorf("subdir/file2.txt not found: %v", err)
	}
}

func TestCollectCmd_SingleFile(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(sourceDir, "single.txt")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Execute collect with single file
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "-c", configDir, sourceFile})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify file was copied
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	destFile := filepath.Join(sheetDir, "single.txt")
	if _, err := os.Stat(destFile); err != nil {
		t.Errorf("single.txt not found: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("content = %q, want %q", string(content), "content")
	}
}

func TestCollectCmd_NonRecursive(t *testing.T) {
	// Setup directories
	configDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(sourceDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// Execute collect with --no-recursive
	cli := NewCLI()
	err := cli.Execute([]string{"collect", "-s", "test-sheet", "--no-recursive", "-c", configDir, sourceDir})

	// Assert
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Verify only top-level file was copied
	sheetDir := filepath.Join(configDir, "sheets", "test-sheet")
	if _, err := os.Stat(filepath.Join(sheetDir, "file1.txt")); err != nil {
		t.Errorf("file1.txt not found: %v", err)
	}

	// Verify subdirectory was not copied
	if _, err := os.Stat(filepath.Join(sheetDir, "subdir")); !os.IsNotExist(err) {
		t.Error("subdir should not exist in non-recursive mode")
	}
	if _, err := os.Stat(filepath.Join(sheetDir, "subdir", "file2.txt")); !os.IsNotExist(err) {
		t.Error("subdir/file2.txt should not exist in non-recursive mode")
	}
}
