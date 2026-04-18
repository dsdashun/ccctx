package runner

import (
	"strings"
	"testing"

	"github.com/dsdashun/ccctx/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildEnv(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "https://old.example.com")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "old-token")
	t.Setenv("ANTHROPIC_MODEL", "old-model")
	t.Setenv("ANTHROPIC_CUSTOM_VAR", "should-be-removed")

	ctx := &config.Context{
		BaseURL:        "https://api.example.com",
		AuthToken:      "secret-token",
		Model:          "claude-sonnet-4-6",
		SmallFastModel: "claude-haiku-4-5",
	}

	env := buildEnv(ctx, Options{})

	assertEnvContains(t, env, "ANTHROPIC_BASE_URL=https://api.example.com")
	assertEnvContains(t, env, "ANTHROPIC_AUTH_TOKEN=secret-token")
	assertEnvContains(t, env, "ANTHROPIC_MODEL=claude-sonnet-4-6")
	assertEnvContains(t, env, "ANTHROPIC_SMALL_FAST_MODEL=claude-haiku-4-5")

	for _, e := range env {
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_CUSTOM_VAR="), "ANTHROPIC_CUSTOM_VAR should have been stripped")
	}
}

func TestBuildEnv_SkipsEmptyOptional(t *testing.T) {
	ctx := &config.Context{
		BaseURL:        "https://api.example.com",
		AuthToken:      "secret-token",
		Model:          "",
		SmallFastModel: "",
	}

	env := buildEnv(ctx, Options{})

	assertEnvContains(t, env, "ANTHROPIC_BASE_URL=https://api.example.com")
	assertEnvContains(t, env, "ANTHROPIC_AUTH_TOKEN=secret-token")
	for _, e := range env {
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_MODEL="), "empty MODEL should not be injected")
		assert.False(t, strings.HasPrefix(e, "ANTHROPIC_SMALL_FAST_MODEL="), "empty SMALL_FAST_MODEL should not be injected")
	}
}

func TestBuildEnv_InjectionOrder(t *testing.T) {
	ctx := &config.Context{
		BaseURL:        "https://api.example.com",
		AuthToken:      "secret-token",
		Model:          "model-val",
		SmallFastModel: "sfm-val",
	}

	env := buildEnv(ctx, Options{})

	var lastIdx int
	for _, prefix := range []string{
		"ANTHROPIC_BASE_URL=",
		"ANTHROPIC_AUTH_TOKEN=",
		"ANTHROPIC_MODEL=",
		"ANTHROPIC_SMALL_FAST_MODEL=",
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

func assertEnvContains(t *testing.T, env []string, expected string) {
	t.Helper()
	for _, e := range env {
		if e == expected {
			return
		}
	}
	t.Errorf("env does not contain %q", expected)
}
