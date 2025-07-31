package config

import (
	"os"
	"testing"
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