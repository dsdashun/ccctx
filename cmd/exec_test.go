package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecRun_ModelFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		configTOML      string
		wantCode        int
		wantModel       string
		wantHaikuModel  string
		wantSonnetModel string
		wantOpusModel   string
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
			name: "--haiku-model without value returns error",
			args:  []string{"--haiku-model"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode: 1,
		},
		{
			name: "--sonnet-model without value returns error",
			args:  []string{"--sonnet-model"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode: 1,
		},
		{
			name: "--opus-model without value returns error",
			args:  []string{"--opus-model"},
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
			name: "--small-fast-model alias sets ANTHROPIC_DEFAULT_HAIKU_MODEL via shell",
			args: []string{"--small-fast-model", "claude-haiku-4-5-20251001", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantHaikuModel: "claude-haiku-4-5-20251001",
		},
		{
			name: "both --model and --small-fast-model set env vars via shell",
			args: []string{"test", "--model", "claude-opus-4-7", "--small-fast-model", "claude-haiku-4-5-20251001"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantModel:      "claude-opus-4-7",
			wantHaikuModel: "claude-haiku-4-5-20251001",
		},
		{
			name: "--haiku-model flag sets ANTHROPIC_DEFAULT_HAIKU_MODEL via shell",
			args: []string{"--haiku-model", "claude-haiku-4-5-20251001", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantHaikuModel: "claude-haiku-4-5-20251001",
		},
		{
			name: "--sonnet-model flag sets ANTHROPIC_DEFAULT_SONNET_MODEL via shell",
			args: []string{"--sonnet-model", "claude-sonnet-4-6", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:        0,
			wantSonnetModel: "claude-sonnet-4-6",
		},
		{
			name: "--opus-model flag sets ANTHROPIC_DEFAULT_OPUS_MODEL via shell",
			args: []string{"--opus-model", "claude-opus-4-7", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:      0,
			wantOpusModel: "claude-opus-4-7",
		},
		{
			name: "--haiku-model wins over --small-fast-model via shell",
			args: []string{"test", "--haiku-model", "X", "--small-fast-model", "Y"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantHaikuModel: "X",
		},
		{
			name: "all model flags combined via shell",
			args: []string{"test", "--model", "m", "--haiku-model", "h", "--sonnet-model", "s", "--opus-model", "o", "--small-fast-model", "sf"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:        0,
			wantModel:       "m",
			wantHaikuModel:  "h",
			wantSonnetModel: "s",
			wantOpusModel:   "o",
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

			hasExplicitTarget := false
			for i, a := range tt.args {
				if a == "--" && i+1 < len(tt.args) {
					hasExplicitTarget = true
					break
				}
			}

			if hasExplicitTarget {
				for i, a := range tt.args {
					if a == "--" && i+1 < len(tt.args) {
						writeModelMock(t, mockDir, tt.args[i+1])
						break
					}
				}
			} else {
				bashPath := filepath.Join(mockDir, "bash")
				writeModelMock(t, mockDir, "bash")
				t.Setenv("SHELL", bashPath)
			}

			t.Setenv("PATH", mockDir)

			code := execRun(tt.args)
			assert.Equal(t, tt.wantCode, code)

			if tt.wantCode == 0 {
				assertModelOutput(t, outputFile, tt.wantModel, tt.wantHaikuModel, tt.wantSonnetModel, tt.wantOpusModel)
			}
		})
	}
}

func TestExecRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		configTOML string
		wantCode   int
	}{
		{
			name: "context not found",
			args: []string{"nonexistent", "--", "env"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode: 1,
		},
		{
			name: "missing base_url in context",
			args: []string{"badctx", "--", "env"},
			configTOML: `[context.badctx]
			auth_token = "test-token"
			`,
			wantCode: 1,
		},
		{
			name: "missing auth_token in context",
			args: []string{"noauthtoken", "--", "env"},
			configTOML: `[context.noauthtoken]
			base_url = "https://api.example.com"
			`,
			wantCode: 1,
		},
		{
			name: "invalid base_url format",
			args: []string{"badurl", "--", "env"},
			configTOML: `[context.badurl]
			base_url = "not-a-valid-url"
			auth_token = "test-token"
			`,
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")
			err := os.WriteFile(configPath, []byte(tt.configTOML), 0600)
			require.NoError(t, err)
			t.Setenv("CCCTX_CONFIG_PATH", configPath)

			code := execRun(tt.args)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}
