package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidYAML(t *testing.T) {
	// Create temp file with valid YAML
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `name: bob
org: example
repo: stamp`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load and verify
	vars, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	expected := map[string]string{
		"name": "bob",
		"org":  "example",
		"repo": "stamp",
	}

	if len(vars) != len(expected) {
		t.Errorf("got %d vars, want %d", len(vars), len(expected))
	}

	for k, want := range expected {
		if got := vars[k]; got != want {
			t.Errorf("vars[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("Load() should return error for non-existent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bad.yaml")
	content := `name: bob
invalid: [unclosed`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() should return error for invalid YAML")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "empty.yaml")

	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	vars, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(vars) != 0 {
		t.Errorf("empty config should return empty map, got %v", vars)
	}
}

func TestLoad_NumberValues(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "numbers.yaml")
	content := `port: 8080
enabled: true
version: 1.5`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	vars, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Numbers and booleans should be stringified
	if vars["port"] != "8080" {
		t.Errorf("vars[port] = %q, want \"8080\"", vars["port"])
	}
	if vars["enabled"] != "true" {
		t.Errorf("vars[enabled] = %q, want \"true\"", vars["enabled"])
	}
	if vars["version"] != "1.5" {
		t.Errorf("vars[version] = %q, want \"1.5\"", vars["version"])
	}
}
