package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modelMockScript = "#!/bin/sh\necho \"$ANTHROPIC_MODEL\" > \"$MOCK_OUTPUT_FILE\"\necho \"$ANTHROPIC_DEFAULT_HAIKU_MODEL\" >> \"$MOCK_OUTPUT_FILE\"\necho \"$ANTHROPIC_DEFAULT_SONNET_MODEL\" >> \"$MOCK_OUTPUT_FILE\"\necho \"$ANTHROPIC_DEFAULT_OPUS_MODEL\" >> \"$MOCK_OUTPUT_FILE\"\nexit 0"

func writeModelMock(t *testing.T, dir, name string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(modelMockScript), 0755)
	require.NoError(t, err)
}

func assertModelOutput(t *testing.T, outputFile, wantModel, wantHaiku, wantSonnet, wantOpus string) {
	t.Helper()
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	lines := strings.Split(string(data), "\n")
	if wantModel != "" {
		require.GreaterOrEqual(t, len(lines), 1, "mock did not write ANTHROPIC_MODEL")
		assert.Equal(t, wantModel, lines[0])
	}
	if wantHaiku != "" {
		require.GreaterOrEqual(t, len(lines), 2, "mock did not write ANTHROPIC_DEFAULT_HAIKU_MODEL")
		assert.Equal(t, wantHaiku, lines[1])
	}
	if wantSonnet != "" {
		require.GreaterOrEqual(t, len(lines), 3, "mock did not write ANTHROPIC_DEFAULT_SONNET_MODEL")
		assert.Equal(t, wantSonnet, lines[2])
	}
	if wantOpus != "" {
		require.GreaterOrEqual(t, len(lines), 4, "mock did not write ANTHROPIC_DEFAULT_OPUS_MODEL")
		assert.Equal(t, wantOpus, lines[3])
	}
}

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
			args: []string{"--unknown-flag", "foo"},
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

func TestRunRun_ModelFlags(t *testing.T) {
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
			name: "--model flag sets ANTHROPIC_MODEL",
			args: []string{"--model", "claude-opus-4-7", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:  0,
			wantModel: "claude-opus-4-7",
		},
		{
			name: "--model with separator forwards args",
			args: []string{"--model", "claude-opus-4-7", "test", "--", "--help"},
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
			name: "both --model and --small-fast-model set env vars",
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
			name: "--haiku-model flag sets ANTHROPIC_DEFAULT_HAIKU_MODEL",
			args: []string{"--haiku-model", "claude-haiku-4-5-20251001", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantHaikuModel: "claude-haiku-4-5-20251001",
		},
		{
			name: "--sonnet-model flag sets ANTHROPIC_DEFAULT_SONNET_MODEL",
			args: []string{"--sonnet-model", "claude-sonnet-4-6", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:        0,
			wantSonnetModel: "claude-sonnet-4-6",
		},
		{
			name: "--opus-model flag sets ANTHROPIC_DEFAULT_OPUS_MODEL",
			args: []string{"--opus-model", "claude-opus-4-7", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:      0,
			wantOpusModel: "claude-opus-4-7",
		},
		{
			name: "--small-fast-model alias sets ANTHROPIC_DEFAULT_HAIKU_MODEL",
			args: []string{"--small-fast-model", "claude-haiku-4-5-20251001", "test"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantHaikuModel: "claude-haiku-4-5-20251001",
		},
		{
			name: "--haiku-model wins over --small-fast-model",
			args: []string{"test", "--haiku-model", "X", "--small-fast-model", "Y"},
			configTOML: `[context.test]
			base_url = "https://api.example.com"
			auth_token = "test-token"
			`,
			wantCode:       0,
			wantHaikuModel: "X",
		},
		{
			name: "all model flags combined",
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
			writeModelMock(t, mockDir, "claude")
			t.Setenv("PATH", mockDir)

			code := runRun(tt.args)
			assert.Equal(t, tt.wantCode, code)

			if tt.wantCode == 0 {
				assertModelOutput(t, outputFile, tt.wantModel, tt.wantHaikuModel, tt.wantSonnetModel, tt.wantOpusModel)
			}
		})
	}
}
