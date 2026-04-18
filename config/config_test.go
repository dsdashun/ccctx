package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveEnvVar(t *testing.T) {
	// Save and restore environment variables
	originalTestToken := os.Getenv("TEST_TOKEN")
	originalEmptyToken := os.Getenv("EMPTY_TOKEN")
	defer func() {
		if originalTestToken != "" {
			os.Setenv("TEST_TOKEN", originalTestToken)
		} else {
			os.Unsetenv("TEST_TOKEN")
		}
		if originalEmptyToken != "" {
			os.Setenv("EMPTY_TOKEN", originalEmptyToken)
		} else {
			os.Unsetenv("EMPTY_TOKEN")
		}
	}()

	// Set up test environment variables
	os.Setenv("TEST_TOKEN", "test-token-value")
	os.Setenv("EMPTY_TOKEN", "")

	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "regular value without env prefix",
			input:   "direct-token-value",
			want:    "direct-token-value",
			wantErr: false,
		},
		{
			name:    "valid environment variable",
			input:   "env:TEST_TOKEN",
			want:    "test-token-value",
			wantErr: false,
		},
		{
			name:        "empty environment variable name",
			input:       "env:",
			wantErr:     true,
			errContains: "environment variable name cannot be empty",
		},
		{
			name:        "missing environment variable",
			input:       "env:MISSING_TOKEN",
			wantErr:     true,
			errContains: "environment variable 'MISSING_TOKEN' is not set or empty",
		},
		{
			name:        "empty environment variable value",
			input:       "env:EMPTY_TOKEN",
			wantErr:     true,
			errContains: "environment variable 'EMPTY_TOKEN' is not set or empty",
		},
		{
			name:    "env prefix in middle of string",
			input:   "some-env:text",
			want:    "some-env:text",
			wantErr: false,
		},
		{
			name:    "value starting with env but not colon",
			input:   "envtext",
			want:    "envtext",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveEnvVar(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveEnvVar() expected error, got nil")
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("resolveEnvVar() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("resolveEnvVar() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("resolveEnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigFilePermissions(t *testing.T) {
	tests := []struct {
		name        string
		createFirst bool
		initialPerm os.FileMode
	}{
		{
			name:        "auto-created file has 0600 permissions",
			createFirst: false,
		},
		{
			name:        "existing file with 0644 is not modified",
			createFirst: true,
			initialPerm: 0644,
		},
		{
			name:        "existing file with 0400 is not modified",
			createFirst: true,
			initialPerm: 0400,
		},
		{
			name:        "existing file with 0666 is not modified",
			createFirst: true,
			initialPerm: 0666,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			var beforePerm os.FileMode
			if tt.createFirst {
				content := []byte("[context.test]\nbase_url = \"https://example.com\"\nauth_token = \"token\"\n")
				err := os.WriteFile(configPath, content, tt.initialPerm)
				require.NoError(t, err)

				info, err := os.Stat(configPath)
				require.NoError(t, err)
				beforePerm = info.Mode().Perm()
			}

			t.Setenv("CCCTX_CONFIG_PATH", configPath)
			_, err := LoadConfig()
			require.NoError(t, err)

			info, err := os.Stat(configPath)
			require.NoError(t, err)

			if tt.createFirst {
				assert.Equal(t, beforePerm, info.Mode().Perm(), "LoadConfig() modified existing file permissions")
			} else {
				assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
			}
		})
	}
}
