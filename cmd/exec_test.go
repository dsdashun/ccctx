package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecRun_ModelFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		configTOML string
		envSetup   func(t *testing.T, mockDir string, outputFile string)
		wantCode   int
		wantModel  string
		wantSFM    string
	}{
		{
			name: "--model flag sets ANTHROPIC_MODEL via shell",
			args: []string{"--model", "claude-opus-4-7", "test"},
			configTOML: `[context.test]
	base_url = "https://api.example.com"
	auth_token = "test-token"
	`,
			wantCode:  0,
			wantModel: "claude-opus-4-7",
		},
		{
			name: "--model without value returns error",
			args: []string{"--model"},
			configTOML: `[context.test]
	base_url = "https://api.example.com"
	auth_token = "test-token"
	`,
			wantCode: 1,
		},
		{
			name: "--small-fast-model without value returns error",
			args: []string{"--small-fast-model"},
			configTOML: `[context.test]
	base_url = "https://api.example.com"
	auth_token = "test-token"
	`,
			wantCode: 1,
		},
		{
			name: "provider with --model and explicit target command",
			args: []string{"test", "--model", "claude-opus-4-7", "--", "env"},
			configTOML: `[context.test]
	base_url = "https://api.example.com"
	auth_token = "test-token"
	`,
			wantCode:  0,
			wantModel: "claude-opus-4-7",
		},
		{
			name: "--small-fast-model flag sets ANTHROPIC_DEFAULT_HAIKU_MODEL via shell",
			args: []string{"--small-fast-model", "claude-sonnet-4-6", "test"},
			configTOML: `[context.test]
	base_url = "https://api.example.com"
	auth_token = "test-token"
	`,
			wantCode: 0,
			wantSFM:  "claude-sonnet-4-6",
		},
		{
			name: "both flags set both env vars via shell",
			args: []string{"test", "--model", "claude-opus-4-7", "--small-fast-model", "claude-sonnet-4-6"},
			configTOML: `[context.test]
	base_url = "https://api.example.com"
	auth_token = "test-token"
	`,
			wantCode:  0,
			wantModel: "claude-opus-4-7",
			wantSFM:   "claude-sonnet-4-6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")
			err := os.WriteFile(configPath, []byte(tt.configTOML), 0600)
			require.NoError(t, err)
			t.Setenv("CCCTX_CONFIG_PATH", configPath)

			outputFile := filepath.Join(t.TempDir(), "mock_output")
			t.Setenv("MOCK_OUTPUT_FILE", outputFile)

			mockDir := t.TempDir()
			mockScript := []byte("#!/bin/sh\necho \"$ANTHROPIC_MODEL\" > \"$MOCK_OUTPUT_FILE\"\necho \"$ANTHROPIC_DEFAULT_HAIKU_MODEL\" >> \"$MOCK_OUTPUT_FILE\"\nexit 0")

			// Determine mock name based on whether args include explicit target after --
			hasExplicitTarget := false
			for i, a := range tt.args {
				if a == "--" && i+1 < len(tt.args) {
					hasExplicitTarget = true
					break
				}
			}

			if hasExplicitTarget {
				// Find the first target arg after --
				for i, a := range tt.args {
					if a == "--" && i+1 < len(tt.args) {
						mockName := tt.args[i+1]
						err = os.WriteFile(filepath.Join(mockDir, mockName), mockScript, 0755)
						require.NoError(t, err)
						break
					}
				}
			} else {
				// exec falls back to $SHELL — mock as "bash"
				err = os.WriteFile(filepath.Join(mockDir, "bash"), mockScript, 0755)
				require.NoError(t, err)
				t.Setenv("SHELL", filepath.Join(mockDir, "bash"))
			}

			t.Setenv("PATH", mockDir)

			code := execRun(tt.args)
			assert.Equal(t, tt.wantCode, code)

			if tt.wantCode == 0 {
				data, err := os.ReadFile(outputFile)
				require.NoError(t, err)
				lines := strings.Split(string(data), "\n")
				if tt.wantModel != "" {
					require.GreaterOrEqual(t, len(lines), 1, "mock did not write ANTHROPIC_MODEL")
					assert.Equal(t, tt.wantModel, lines[0])
				}
				if tt.wantSFM != "" {
					require.GreaterOrEqual(t, len(lines), 2, "mock did not write ANTHROPIC_DEFAULT_HAIKU_MODEL")
					assert.Equal(t, tt.wantSFM, lines[1])
				}
			}
		})
	}
}
