package runner

import (
	"strings"
	"testing"

	"github.com/dsdashun/ccctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEnv(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "https://old.example.com")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "old-token")
	t.Setenv("ANTHROPIC_MODEL", "old-model")
	t.Setenv("ANTHROPIC_CUSTOM_VAR", "should-be-removed")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "should-be-stripped")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "should-be-stripped")
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "should-be-stripped")

	ctx := &config.Context{
		BaseURL:     "https://api.example.com",
		AuthToken:   "secret-token",
		Model:       "claude-sonnet-4-6",
		HaikuModel:  "claude-haiku-4-5",
		SonnetModel: "claude-sonnet-4-6",
		OpusModel:   "claude-opus-4-7",
	}

	env := buildEnv(ctx, Options{})

	assertEnvContains(t, env, "ANTHROPIC_BASE_URL=https://api.example.com")
	assertEnvContains(t, env, "ANTHROPIC_AUTH_TOKEN=secret-token")
	assertEnvContains(t, env, "ANTHROPIC_MODEL=claude-sonnet-4-6")
	assertEnvContains(t, env, "ANTHROPIC_DEFAULT_HAIKU_MODEL=claude-haiku-4-5")
	assertEnvContains(t, env, "ANTHROPIC_DEFAULT_SONNET_MODEL=claude-sonnet-4-6")
	assertEnvContains(t, env, "ANTHROPIC_DEFAULT_OPUS_MODEL=claude-opus-4-7")

	for _, e := range env {
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_CUSTOM_VAR="), "ANTHROPIC_CUSTOM_VAR should have been stripped")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_SMALL_FAST_MODEL="), "ANTHROPIC_SMALL_FAST_MODEL should never be injected")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_HAIKU_MODEL=should-be-stripped"), "ANTHROPIC_DEFAULT_HAIKU_MODEL should have been stripped")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_SONNET_MODEL=should-be-stripped"), "ANTHROPIC_DEFAULT_SONNET_MODEL should have been stripped")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_OPUS_MODEL=should-be-stripped"), "ANTHROPIC_DEFAULT_OPUS_MODEL should have been stripped")
	}
}

func TestBuildEnv_SkipsEmptyOptional(t *testing.T) {
	ctx := &config.Context{
		BaseURL:        "https://api.example.com",
		AuthToken:      "secret-token",
		Model:          "",
		SmallFastModel: "",
		HaikuModel:     "",
		SonnetModel:    "",
		OpusModel:      "",
	}

	env := buildEnv(ctx, Options{})

	assertEnvContains(t, env, "ANTHROPIC_BASE_URL=https://api.example.com")
	assertEnvContains(t, env, "ANTHROPIC_AUTH_TOKEN=secret-token")
	for _, e := range env {
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_MODEL="), "empty MODEL should not be injected")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_SMALL_FAST_MODEL="), "ANTHROPIC_SMALL_FAST_MODEL should never be injected")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_HAIKU_MODEL="), "empty HAIKU should not be injected")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_SONNET_MODEL="), "empty SONNET should not be injected")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_OPUS_MODEL="), "empty OPUS should not be injected")
	}
}

func TestBuildEnv_InjectionOrder(t *testing.T) {
	ctx := &config.Context{
		BaseURL:     "https://api.example.com",
		AuthToken:   "secret-token",
		Model:       "model-val",
		HaikuModel:  "haiku-val",
		SonnetModel: "sonnet-val",
		OpusModel:   "opus-val",
	}

	env := buildEnv(ctx, Options{})

	var lastIdx int
	for _, prefix := range []string{
		"ANTHROPIC_BASE_URL=",
		"ANTHROPIC_AUTH_TOKEN=",
		"ANTHROPIC_MODEL=",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL=",
		"ANTHROPIC_DEFAULT_SONNET_MODEL=",
		"ANTHROPIC_DEFAULT_OPUS_MODEL=",
	} {
		idx := -1
		for i, e := range env {
			if strings.HasPrefix(e, prefix) {
				idx = i
				break
			}
		}
		assert.GreaterOrEqual(t, idx, 0, "expected to find %s in env", prefix)
		assert.GreaterOrEqual(t, idx, lastIdx, "%s should appear after previous injection", prefix)
		lastIdx = idx
	}
}

func TestBuildEnv_PriorityChain(t *testing.T) {
	tests := []struct {
		name      string
		ctxModel  string
		optsModel string
		wantModel string
	}{
		{
			name:      "opts.Model set, ctx.Model empty → uses opts.Model",
			ctxModel:  "",
			optsModel: "claude-opus-4-7",
			wantModel: "claude-opus-4-7",
		},
		{
			name:      "opts.Model set, ctx.Model set → opts wins",
			ctxModel:  "claude-sonnet-4-6",
			optsModel: "claude-opus-4-7",
			wantModel: "claude-opus-4-7",
		},
		{
			name:      "opts.Model empty, ctx.Model set → uses ctx.Model",
			ctxModel:  "claude-sonnet-4-6",
			optsModel: "",
			wantModel: "claude-sonnet-4-6",
		},
		{
			name:      "both empty → no ANTHROPIC_MODEL",
			ctxModel:  "",
			optsModel: "",
			wantModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &config.Context{
				BaseURL:   "https://api.example.com",
				AuthToken: "test-token",
				Model:     tt.ctxModel,
			}
			opts := Options{
				Model: tt.optsModel,
			}

			env := buildEnv(ctx, opts)

			if tt.wantModel != "" {
				assertEnvContains(t, env, "ANTHROPIC_MODEL="+tt.wantModel)
			} else {
				for _, e := range env {
					assert.False(t, strings.HasPrefix(e, "ANTHROPIC_MODEL="), "expected no ANTHROPIC_MODEL, got %q", e)
				}
			}
		})
	}
}

func TestBuildEnv_HaikuPriorityChain(t *testing.T) {
	tests := []struct {
		name      string
		ctxHaiku  string
		ctxSFM    string
		optsHaiku string
		optsSFM   string
		wantHaiku string
	}{
		{
			name:      "opts.HaikuModel set → uses it regardless of other fields",
			ctxHaiku:  "ctx-haiku",
			ctxSFM:    "ctx-sfm",
			optsHaiku: "opts-haiku",
			optsSFM:   "opts-sfm",
			wantHaiku: "opts-haiku",
		},
		{
			name:      "opts.HaikuModel empty, opts.SFM set → uses SFM as haiku",
			ctxHaiku:  "ctx-haiku",
			ctxSFM:    "ctx-sfm",
			optsHaiku: "",
			optsSFM:   "opts-sfm",
			wantHaiku: "opts-sfm",
		},
		{
			name:      "opts.HaikuModel empty, opts.SFM empty, ctx.HaikuModel set → uses ctx.HaikuModel",
			ctxHaiku:  "ctx-haiku",
			ctxSFM:    "ctx-sfm",
			optsHaiku: "",
			optsSFM:   "",
			wantHaiku: "ctx-haiku",
		},
		{
			name:      "opts.HaikuModel empty, opts.SFM empty, ctx.HaikuModel empty, ctx.SFM set → uses ctx.SFM",
			ctxHaiku:  "",
			ctxSFM:    "ctx-sfm",
			optsHaiku: "",
			optsSFM:   "",
			wantHaiku: "ctx-sfm",
		},
		{
			name:      "all empty → no ANTHROPIC_DEFAULT_HAIKU_MODEL",
			wantHaiku: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &config.Context{
				BaseURL:        "https://api.example.com",
				AuthToken:      "test-token",
				HaikuModel:     tt.ctxHaiku,
				SmallFastModel: tt.ctxSFM,
			}
			opts := Options{
				HaikuModel:     tt.optsHaiku,
				SmallFastModel: tt.optsSFM,
			}

			env := buildEnv(ctx, opts)

			if tt.wantHaiku != "" {
				assertEnvContains(t, env, "ANTHROPIC_DEFAULT_HAIKU_MODEL="+tt.wantHaiku)
			} else {
				for _, e := range env {
					assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_HAIKU_MODEL="), "expected no ANTHROPIC_DEFAULT_HAIKU_MODEL, got %q", e)
				}
			}
			for _, e := range env {
				assert.False(t, strings.HasPrefix(e, "ANTHROPIC_SMALL_FAST_MODEL="), "ANTHROPIC_SMALL_FAST_MODEL should never be injected")
			}
		})
	}
}

func TestBuildEnv_SonnetOpusPriorityChain(t *testing.T) {
	tests := []struct {
		name       string
		ctxSonnet  string
		ctxOpus    string
		optsSonnet string
		optsOpus   string
		wantSonnet string
		wantOpus   string
	}{
		{
			name:       "opts.SonnetModel wins over ctx.SonnetModel",
			ctxSonnet:  "ctx-sonnet",
			optsSonnet: "opts-sonnet",
			wantSonnet: "opts-sonnet",
		},
		{
			name:       "opts.SonnetModel empty, ctx.SonnetModel set → uses ctx",
			ctxSonnet:  "ctx-sonnet",
			optsSonnet: "",
			wantSonnet: "ctx-sonnet",
		},
		{
			name:       "both sonnet empty → no ANTHROPIC_DEFAULT_SONNET_MODEL",
			wantSonnet: "",
		},
		{
			name:     "opts.OpusModel wins over ctx.OpusModel",
			ctxOpus:  "ctx-opus",
			optsOpus: "opts-opus",
			wantOpus: "opts-opus",
		},
		{
			name:     "opts.OpusModel empty, ctx.OpusModel set → uses ctx",
			ctxOpus:  "ctx-opus",
			optsOpus: "",
			wantOpus: "ctx-opus",
		},
		{
			name:     "both opus empty → no ANTHROPIC_DEFAULT_OPUS_MODEL",
			wantOpus: "",
		},
		{
			name:       "both sonnet and opus set → both injected",
			ctxSonnet:  "ctx-sonnet",
			ctxOpus:    "ctx-opus",
			optsSonnet: "",
			optsOpus:   "",
			wantSonnet: "ctx-sonnet",
			wantOpus:   "ctx-opus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &config.Context{
				BaseURL:     "https://api.example.com",
				AuthToken:   "test-token",
				SonnetModel: tt.ctxSonnet,
				OpusModel:   tt.ctxOpus,
			}
			opts := Options{
				SonnetModel: tt.optsSonnet,
				OpusModel:   tt.optsOpus,
			}

			env := buildEnv(ctx, opts)

			if tt.wantSonnet != "" {
				assertEnvContains(t, env, "ANTHROPIC_DEFAULT_SONNET_MODEL="+tt.wantSonnet)
			} else {
				for _, e := range env {
					assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_SONNET_MODEL="), "expected no ANTHROPIC_DEFAULT_SONNET_MODEL, got %q", e)
				}
			}

			if tt.wantOpus != "" {
				assertEnvContains(t, env, "ANTHROPIC_DEFAULT_OPUS_MODEL="+tt.wantOpus)
			} else {
				for _, e := range env {
					assert.False(t, strings.HasPrefix(e, "ANTHROPIC_DEFAULT_OPUS_MODEL="), "expected no ANTHROPIC_DEFAULT_OPUS_MODEL, got %q", e)
				}
			}
		})
	}
}

func TestBuildEnv_SmallFastModelNotInjected(t *testing.T) {
	tests := []struct {
		name      string
		ctxSFM    string
		optsSFM   string
		wantHaiku string
	}{
		{
			name:      "ctx.SmallFastModel set but ANTHROPIC_SMALL_FAST_MODEL never injected",
			ctxSFM:    "claude-haiku-4-5",
			wantHaiku: "claude-haiku-4-5",
		},
		{
			name:      "opts.SmallFastModel set but ANTHROPIC_SMALL_FAST_MODEL never injected",
			optsSFM:   "claude-haiku-4-5",
			wantHaiku: "claude-haiku-4-5",
		},
		{
			name:      "both ctx and opts SFM set but ANTHROPIC_SMALL_FAST_MODEL never injected",
			ctxSFM:    "ctx-sfm",
			optsSFM:   "opts-sfm",
			wantHaiku: "opts-sfm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &config.Context{
				BaseURL:        "https://api.example.com",
				AuthToken:      "test-token",
				SmallFastModel: tt.ctxSFM,
			}
			opts := Options{
				SmallFastModel: tt.optsSFM,
			}

			env := buildEnv(ctx, opts)

			assertEnvContains(t, env, "ANTHROPIC_DEFAULT_HAIKU_MODEL="+tt.wantHaiku)
			for _, e := range env {
				assert.False(t, strings.HasPrefix(e, "ANTHROPIC_SMALL_FAST_MODEL="), "ANTHROPIC_SMALL_FAST_MODEL should never be injected, got %q", e)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr string
	}{
		{
			name:   "valid https URL",
			rawURL: "https://api.example.com",
		},
		{
			name:   "valid http URL",
			rawURL: "http://localhost:8080",
		},
		{
			name:    "missing scheme",
			rawURL:  "api.example.com",
			wantErr: "missing scheme",
		},
		{
			name:    "contains spaces",
			rawURL:  "https://api example.com",
			wantErr: "contains spaces",
		},
		{
			name:    "empty string",
			rawURL:  "",
			wantErr: "missing scheme",
		},
		{
			name:   "valid URL with path",
			rawURL: "https://api.example.com/v1",
		},
		{
			name:   "valid URL with port",
			rawURL: "https://api.example.com:8443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.rawURL)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func assertEnvContains(t *testing.T, env []string, expected string) {
	t.Helper()
	for _, e := range env {
		if e == expected {
			return
		}
	}
	t.Errorf("env does not contain %q", expected)
}
