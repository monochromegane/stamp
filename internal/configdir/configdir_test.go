package configdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name          string
		xdgConfigHome string
		wantContains  string
		wantErr       bool
		setupXDG      bool
		clearXDG      bool
	}{
		{
			name:          "with XDG_CONFIG_HOME set",
			xdgConfigHome: "/custom/config",
			wantContains:  "/custom/config/stamp",
			setupXDG:      true,
		},
		{
			name:         "without XDG_CONFIG_HOME (uses UserConfigDir)",
			wantContains: "/stamp",
			clearXDG:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

			if tt.setupXDG {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else if tt.clearXDG {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			got, err := GetConfigDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(filepath.ToSlash(got), tt.wantContains) {
				t.Errorf("GetConfigDir() = %v, want to contain %v", got, tt.wantContains)
			}
		})
	}
}

func TestGetConfigDirWithOverride(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	validDir := filepath.Join(tmpDir, "valid-config")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a file (not a directory) for testing
	notADir := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(notADir, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		override     string
		wantContains string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "empty override uses default",
			override:     "",
			wantContains: "/stamp",
			wantErr:      false,
		},
		{
			name:         "valid override directory",
			override:     validDir,
			wantContains: validDir,
			wantErr:      false,
		},
		{
			name:        "non-existent override directory",
			override:    filepath.Join(tmpDir, "does-not-exist"),
			wantErr:     true,
			errContains: "config directory not found",
		},
		{
			name:        "override is a file, not directory",
			override:    notADir,
			wantErr:     true,
			errContains: "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfigDirWithOverride(tt.override)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigDirWithOverride() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetConfigDirWithOverride() error = %v, want to contain %v", err, tt.errContains)
				}
			} else {
				if !strings.Contains(filepath.ToSlash(got), filepath.ToSlash(tt.wantContains)) {
					t.Errorf("GetConfigDirWithOverride() = %v, want to contain %v", got, tt.wantContains)
				}
			}
		})
	}
}

func TestResolveTemplateDir(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	sheetsDir := filepath.Join(tmpDir, "sheets")
	template1 := filepath.Join(sheetsDir, "go-cli")
	template2 := filepath.Join(sheetsDir, "web-app")

	if err := os.MkdirAll(template1, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(template2, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a file in sheets dir (not a directory)
	notADirTemplate := filepath.Join(sheetsDir, "not-a-dir")
	if err := os.WriteFile(notADirTemplate, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		configDir    string
		templateName string
		want         string
		wantErr      bool
		errContains  []string
	}{
		{
			name:         "existing template",
			configDir:    tmpDir,
			templateName: "go-cli",
			want:         template1,
			wantErr:      false,
		},
		{
			name:         "non-existent template with available sheets",
			configDir:    tmpDir,
			templateName: "does-not-exist",
			wantErr:      true,
			errContains:  []string{"sheet 'does-not-exist' not found", "Available sheets:", "go-cli", "web-app"},
		},
		{
			name:         "non-existent template without sheets directory",
			configDir:    t.TempDir(), // Empty temp dir
			templateName: "any-template",
			wantErr:      true,
			errContains:  []string{"sheet 'any-template' not found", "Create sheet directory"},
		},
		{
			name:         "template path is a file, not directory",
			configDir:    tmpDir,
			templateName: "not-a-dir",
			wantErr:      true,
			errContains:  []string{"not a directory"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveTemplateDir(tt.configDir, tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTemplateDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				errMsg := err.Error()
				for _, contains := range tt.errContains {
					if !strings.Contains(errMsg, contains) {
						t.Errorf("ResolveTemplateDir() error = %v, want to contain %v", errMsg, contains)
					}
				}
			} else {
				if got != tt.want {
					t.Errorf("ResolveTemplateDir() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestListAvailableSheets(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	sheetsDir := filepath.Join(tmpDir, "sheets")

	// Test with sheets directory
	template1 := filepath.Join(sheetsDir, "go-cli")
	template2 := filepath.Join(sheetsDir, "web-app")
	template3 := filepath.Join(sheetsDir, "api-service")

	if err := os.MkdirAll(template1, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(template2, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(template3, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a file in sheets dir (should be ignored)
	if err := os.WriteFile(filepath.Join(sheetsDir, "readme.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		configDir string
		want      []string
		wantErr   bool
	}{
		{
			name:      "with sheets",
			configDir: tmpDir,
			want:      []string{"api-service", "go-cli", "web-app"}, // sorted
			wantErr:   false,
		},
		{
			name:      "without sheets directory",
			configDir: t.TempDir(), // Empty temp dir
			want:      []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListAvailableSheets(tt.configDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAvailableSheets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ListAvailableSheets() = %v, want %v", got, tt.want)
					return
				}
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("ListAvailableSheets()[%d] = %v, want %v", i, v, tt.want[i])
					}
				}
			}
		})
	}
}
