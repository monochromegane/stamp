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

func TestLoadHierarchical_BothConfigs(t *testing.T) {
	dir := t.TempDir()

	// Create global config
	globalPath := filepath.Join(dir, "stamp.yaml")
	globalContent := `org: global-org
author: alice
license: MIT`
	if err := os.WriteFile(globalPath, []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Create template-specific config
	templateDir := filepath.Join(dir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	templatePath := filepath.Join(templateDir, "stamp.yaml")
	templateContent := `name: myproject
org: template-org`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template config: %v", err)
	}

	// Load hierarchical configs
	vars, err := LoadHierarchical(dir, "go-cli")
	if err != nil {
		t.Fatalf("LoadHierarchical() failed: %v", err)
	}

	// Verify merge with correct priority (template > global)
	expected := map[string]string{
		"name":    "myproject",    // from template
		"org":     "template-org", // template overrides global
		"author":  "alice",        // from global
		"license": "MIT",          // from global
	}

	if len(vars) != len(expected) {
		t.Errorf("got %d vars, want %d: %v", len(vars), len(expected), vars)
	}

	for k, want := range expected {
		if got := vars[k]; got != want {
			t.Errorf("vars[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestLoadHierarchical_OnlyGlobalConfig(t *testing.T) {
	dir := t.TempDir()

	// Create only global config
	globalPath := filepath.Join(dir, "stamp.yaml")
	globalContent := `org: global-org
author: alice`
	if err := os.WriteFile(globalPath, []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Create template directory but no config
	templateDir := filepath.Join(dir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Load hierarchical configs
	vars, err := LoadHierarchical(dir, "go-cli")
	if err != nil {
		t.Fatalf("LoadHierarchical() failed: %v", err)
	}

	// Should have only global values
	expected := map[string]string{
		"org":    "global-org",
		"author": "alice",
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

func TestLoadHierarchical_OnlyTemplateConfig(t *testing.T) {
	dir := t.TempDir()

	// Create only template-specific config (no global)
	templateDir := filepath.Join(dir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	templatePath := filepath.Join(templateDir, "stamp.yaml")
	templateContent := `name: myproject
org: template-org`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template config: %v", err)
	}

	// Load hierarchical configs
	vars, err := LoadHierarchical(dir, "go-cli")
	if err != nil {
		t.Fatalf("LoadHierarchical() failed: %v", err)
	}

	// Should have only template values
	expected := map[string]string{
		"name": "myproject",
		"org":  "template-org",
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

func TestLoadHierarchical_NeitherConfig(t *testing.T) {
	dir := t.TempDir()

	// Create template directory but no configs
	templateDir := filepath.Join(dir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Load hierarchical configs
	vars, err := LoadHierarchical(dir, "go-cli")
	if err != nil {
		t.Fatalf("LoadHierarchical() should not error when no configs exist: %v", err)
	}

	// Should return empty map
	if len(vars) != 0 {
		t.Errorf("got %d vars, want 0 (empty map)", len(vars))
	}
}

func TestLoadHierarchical_InvalidGlobalConfig(t *testing.T) {
	dir := t.TempDir()

	// Create invalid global config
	globalPath := filepath.Join(dir, "stamp.yaml")
	globalContent := `invalid: [unclosed`
	if err := os.WriteFile(globalPath, []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Load hierarchical configs
	_, err := LoadHierarchical(dir, "go-cli")
	if err == nil {
		t.Fatal("LoadHierarchical() should return error for invalid global config")
	}
}

func TestLoadHierarchical_InvalidTemplateConfig(t *testing.T) {
	dir := t.TempDir()

	// Create template with invalid config
	templateDir := filepath.Join(dir, "templates", "go-cli")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	templatePath := filepath.Join(templateDir, "stamp.yaml")
	templateContent := `invalid: [unclosed`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template config: %v", err)
	}

	// Load hierarchical configs
	_, err := LoadHierarchical(dir, "go-cli")
	if err == nil {
		t.Fatal("LoadHierarchical() should return error for invalid template config")
	}
}

func TestLoadOptional_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `name: bob`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	vars, err := loadOptional(configPath)
	if err != nil {
		t.Fatalf("loadOptional() failed: %v", err)
	}

	if vars["name"] != "bob" {
		t.Errorf("vars[name] = %q, want \"bob\"", vars["name"])
	}
}

func TestLoadOptional_NonExistentFile(t *testing.T) {
	vars, err := loadOptional("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("loadOptional() should not error for non-existent file: %v", err)
	}

	if len(vars) != 0 {
		t.Errorf("loadOptional() should return empty map for non-existent file, got %v", vars)
	}
}

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		want     map[string]string
	}{
		{
			name:     "empty base and override",
			base:     map[string]string{},
			override: map[string]string{},
			want:     map[string]string{},
		},
		{
			name: "only base values",
			base: map[string]string{
				"name": "alice",
				"org":  "example",
			},
			override: map[string]string{},
			want: map[string]string{
				"name": "alice",
				"org":  "example",
			},
		},
		{
			name: "only override values",
			base: map[string]string{},
			override: map[string]string{
				"name": "bob",
				"repo": "myrepo",
			},
			want: map[string]string{
				"name": "bob",
				"repo": "myrepo",
			},
		},
		{
			name: "override takes precedence",
			base: map[string]string{
				"name":   "alice",
				"org":    "base-org",
				"author": "alice",
			},
			override: map[string]string{
				"name": "bob",
				"org":  "override-org",
			},
			want: map[string]string{
				"name":   "bob",          // overridden
				"org":    "override-org", // overridden
				"author": "alice",        // from base
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeConfigs(tt.base, tt.override)

			if len(got) != len(tt.want) {
				t.Errorf("got %d vars, want %d", len(got), len(tt.want))
			}

			for k, want := range tt.want {
				if gotVal := got[k]; gotVal != want {
					t.Errorf("result[%q] = %q, want %q", k, gotVal, want)
				}
			}
		})
	}
}
