package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		configTOML string
		envSetup   func(t *testing.T)
		wantCode   int
	}{
		{
			name: "ParseArgs error - flag-like arg",
			args: []string{"--model", "foo"},
			configTOML: `[context.test]
base_url = "https://api.example.com"
auth_token = "test-token"
`,
			wantCode: 1,
		},
		{
			name: "context not found",
			args: []string{"nonexistent"},
			configTOML: `[context.test]
base_url = "https://api.example.com"
auth_token = "test-token"
`,
			wantCode: 1,
		},
		{
			name: "claude not found in PATH",
			args: []string{"test"},
			configTOML: `[context.test]
base_url = "https://api.example.com"
auth_token = "test-token"
`,
			envSetup: func(t *testing.T) {
				t.Setenv("PATH", "/nonexistent")
			},
			wantCode: 1,
		},
		{
			name: "success - provider found and claude found",
			args: []string{"test"},
			configTOML: `[context.test]
base_url = "https://api.example.com"
auth_token = "test-token"
`,
			envSetup: func(t *testing.T) {
				tmpDir := t.TempDir()
				claudePath := filepath.Join(tmpDir, "claude")
				err := os.WriteFile(claudePath, []byte("#!/bin/sh\nexit 0"), 0755)
				require.NoError(t, err)
				t.Setenv("PATH", tmpDir)
			},
			wantCode: 0,
		},
		{
			name: "success - claude exits with non-zero code",
			args: []string{"test"},
			configTOML: `[context.test]
base_url = "https://api.example.com"
auth_token = "test-token"
`,
			envSetup: func(t *testing.T) {
				tmpDir := t.TempDir()
				claudePath := filepath.Join(tmpDir, "claude")
				err := os.WriteFile(claudePath, []byte("#!/bin/sh\nexit 42"), 0755)
				require.NoError(t, err)
				t.Setenv("PATH", tmpDir)
			},
			wantCode: 42,
		},
		{
			name: "missing base_url in context",
			args: []string{"badctx"},
			configTOML: `[context.badctx]
auth_token = "test-token"
`,
			wantCode: 1,
		},
		{
			name: "invalid base_url format",
			args: []string{"badurl"},
			configTOML: `[context.badurl]
base_url = "not-a-valid-url"
auth_token = "test-token"
`,
			wantCode: 1,
		},
		{
			name: "missing auth_token in context",
			args: []string{"noauthtoken"},
			configTOML: `[context.noauthtoken]
base_url = "https://api.example.com"
`,
			wantCode: 1,
		},
		{
			name: "success - provider with forwarded args after separator",
			args: []string{"test", "--", "--model", "foo"},
			configTOML: `[context.test]
base_url = "https://api.example.com"
auth_token = "test-token"
`,
			envSetup: func(t *testing.T) {
				tmpDir := t.TempDir()
				claudePath := filepath.Join(tmpDir, "claude")
				err := os.WriteFile(claudePath, []byte("#!/bin/sh\nexit 0"), 0755)
				require.NoError(t, err)
				t.Setenv("PATH", tmpDir)
			},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")
			err := os.WriteFile(configPath, []byte(tt.configTOML), 0600)
			require.NoError(t, err)
			t.Setenv("CCCTX_CONFIG_PATH", configPath)

			if tt.envSetup != nil {
				tt.envSetup(t)
			}

			code := runRun(tt.args)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}
