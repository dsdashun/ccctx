package runner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantModel       string
		wantHaikuModel  string
		wantSonnetModel string
		wantOpusModel   string
		wantRemaining   []string
		wantErr         string
	}{
		{
			name:            "no flags returns empty strings and args unchanged",
			args:            []string{"provider-A"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--model before provider",
			args:            []string{"--model", "foo", "provider-A"},
			wantModel:       "foo",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--model after provider",
			args:            []string{"provider-A", "--model", "foo"},
			wantModel:       "foo",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--model with separator preserves forwarded args",
			args:            []string{"provider-A", "--model", "foo", "--", "--help"},
			wantModel:       "foo",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A", "--", "--help"},
		},
		{
			name:            "both flags extracted",
			args:            []string{"provider-A", "--model", "foo", "--small-fast-model", "bar", "--", "cmd"},
			wantModel:       "foo",
			wantHaikuModel:  "bar",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A", "--", "cmd"},
		},
		{
			name:          "--model without value returns error",
			args:          []string{"--model"},
			wantErr:       "--model requires a value",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--small-fast-model without value returns error",
			args:          []string{"--small-fast-model"},
			wantErr:       "--small-fast-model requires a value",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:            "--model after separator is NOT extracted",
			args:            []string{"provider-A", "--", "--model", "foo"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A", "--", "--model", "foo"},
		},
		{
			name:            "no provider with flag",
			args:            []string{"--model", "foo"},
			wantModel:       "foo",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{},
		},
		{
			name:            "flag value is dash",
			args:            []string{"provider-A", "--model", "-"},
			wantModel:       "-",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "duplicate flags last wins",
			args:            []string{"provider-A", "--model", "foo", "--model", "bar"},
			wantModel:       "bar",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "duplicate --small-fast-model last wins",
			args:            []string{"provider-A", "--small-fast-model", "foo", "--small-fast-model", "bar"},
			wantModel:       "",
			wantHaikuModel:  "bar",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "flag value looks like another flag",
			args:            []string{"provider-A", "--model", "--small-fast-model"},
			wantModel:       "--small-fast-model",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "separator with no following args",
			args:            []string{"--model", "foo", "--"},
			wantModel:       "foo",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"--"},
		},
		{
			name:          "--model value with newline returns error",
			args:          []string{"provider-A", "--model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--small-fast-model value with newline returns error",
			args:          []string{"provider-A", "--small-fast-model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},

		// New flag tests
		{
			name:            "--haiku-model flag works",
			args:            []string{"provider-A", "--haiku-model", "claude-haiku-4-5-20251001"},
			wantModel:       "",
			wantHaikuModel:  "claude-haiku-4-5-20251001",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--sonnet-model flag works",
			args:            []string{"provider-A", "--sonnet-model", "claude-sonnet-4-6"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "claude-sonnet-4-6",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--opus-model flag works",
			args:            []string{"provider-A", "--opus-model", "claude-opus-4-7"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "claude-opus-4-7",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--small-fast-model sets haikuModel (alias)",
			args:            []string{"provider-A", "--small-fast-model", "claude-haiku-4-5-20251001"},
			wantModel:       "",
			wantHaikuModel:  "claude-haiku-4-5-20251001",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--haiku-model wins over --small-fast-model",
			args:            []string{"provider-A", "--haiku-model", "X", "--small-fast-model", "Y"},
			wantModel:       "",
			wantHaikuModel:  "X",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:          "--haiku-model without value returns error",
			args:          []string{"--haiku-model"},
			wantErr:       "--haiku-model requires a value",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--sonnet-model without value returns error",
			args:          []string{"--sonnet-model"},
			wantErr:       "--sonnet-model requires a value",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--opus-model without value returns error",
			args:          []string{"--opus-model"},
			wantErr:       "--opus-model requires a value",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--haiku-model value with newline returns error",
			args:          []string{"provider-A", "--haiku-model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--sonnet-model value with newline returns error",
			args:          []string{"provider-A", "--sonnet-model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:          "--opus-model value with newline returns error",
			args:          []string{"provider-A", "--opus-model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining: []string{},
		},
		{
			name:            "--haiku-model after separator NOT extracted",
			args:            []string{"provider-A", "--", "--haiku-model", "foo"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A", "--", "--haiku-model", "foo"},
		},
		{
			name:            "--sonnet-model after separator NOT extracted",
			args:            []string{"provider-A", "--", "--sonnet-model", "foo"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A", "--", "--sonnet-model", "foo"},
		},
		{
			name:            "--opus-model after separator NOT extracted",
			args:            []string{"provider-A", "--", "--opus-model", "foo"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A", "--", "--opus-model", "foo"},
		},
		{
			name:            "duplicate --haiku-model last wins",
			args:            []string{"provider-A", "--haiku-model", "foo", "--haiku-model", "bar"},
			wantModel:       "",
			wantHaikuModel:  "bar",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "duplicate --sonnet-model last wins",
			args:            []string{"provider-A", "--sonnet-model", "foo", "--sonnet-model", "bar"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "bar",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "duplicate --opus-model last wins",
			args:            []string{"provider-A", "--opus-model", "foo", "--opus-model", "bar"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "bar",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "all 5 flags combined",
			args:            []string{"provider-A", "--model", "m", "--haiku-model", "h", "--sonnet-model", "s", "--opus-model", "o", "--small-fast-model", "sf"},
			wantModel:       "m",
			wantHaikuModel:  "h",
			wantSonnetModel: "s",
			wantOpusModel:   "o",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "nil args returns empty values",
			args:            nil,
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{},
		},
		{
			name:            "empty args returns empty values",
			args:            []string{},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{},
		},
		{
			name:           "--haiku-model at end of args with no value",
			args:           []string{"provider-A", "--haiku-model"},
			wantErr:        "--haiku-model requires a value",
			wantModel:      "",
			wantHaikuModel: "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:  []string{},
		},
		{
			name:           "--sonnet-model at end of args with no value",
			args:           []string{"provider-A", "--sonnet-model"},
			wantErr:        "--sonnet-model requires a value",
			wantModel:      "",
			wantHaikuModel: "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:  []string{},
		},
		{
			name:           "--opus-model at end of args with no value",
			args:           []string{"provider-A", "--opus-model"},
			wantErr:        "--opus-model requires a value",
			wantModel:      "",
			wantHaikuModel: "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:  []string{},
		},
		{
			name:            "--haiku-model wins over --small-fast-model (reverse order)",
			args:            []string{"provider-A", "--small-fast-model", "Y", "--haiku-model", "X"},
			wantModel:       "",
			wantHaikuModel:  "X",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--haiku-model consumes --sonnet-model as value",
			args:            []string{"provider-A", "--haiku-model", "--sonnet-model"},
			wantModel:       "",
			wantHaikuModel:  "--sonnet-model",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--sonnet-model consumes --opus-model as value",
			args:            []string{"provider-A", "--sonnet-model", "--opus-model"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "--opus-model",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:            "--opus-model consumes --model as value",
			args:            []string{"provider-A", "--opus-model", "--model"},
			wantModel:       "",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "--model",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:           "--haiku-model explicit empty wins over --small-fast-model",
			args:           []string{"provider-A", "--haiku-model", "", "--small-fast-model", "X"},
			wantModel:      "",
			wantHaikuModel: "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:  []string{"provider-A"},
		},
		{
			name:           "--haiku-model explicit empty wins over --small-fast-model (reverse order)",
			args:           []string{"provider-A", "--small-fast-model", "X", "--haiku-model", ""},
			wantModel:      "",
			wantHaikuModel: "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:  []string{"provider-A"},
		},
		{
			name:           "--model with unicode value",
			args:           []string{"provider-A", "--model", "claude-日本語"},
			wantModel:      "claude-日本語",
			wantHaikuModel:  "",
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:           "--haiku-model with 1000+ character value",
			args:           []string{"provider-A", "--haiku-model", strings.Repeat("a", 1001)},
			wantModel:      "",
			wantHaikuModel:  strings.Repeat("a", 1001),
			wantSonnetModel: "",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
		{
			name:           "--sonnet-model with unicode and special characters",
			args:           []string{"provider-A", "--sonnet-model", "claude-日本語-çoğüşiöü"},
			wantModel:      "",
			wantHaikuModel:  "",
			wantSonnetModel: "claude-日本語-çoğüşiöü",
			wantOpusModel:   "",
			wantRemaining:   []string{"provider-A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, haiku, sonnet, opus, remaining, err := ExtractFlags(tt.args)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Equal(t, []string{}, remaining)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantModel, model)
			assert.Equal(t, tt.wantHaikuModel, haiku)
			assert.Equal(t, tt.wantSonnetModel, sonnet)
			assert.Equal(t, tt.wantOpusModel, opus)
			assert.Equal(t, tt.wantRemaining, remaining)
		})
	}
}

func TestWantsHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "--help only",
			args: []string{"--help"},
			want: true,
		},
		{
			name: "-h only",
			args: []string{"-h"},
			want: true,
		},
		{
			name: "--help after separator is not help",
			args: []string{"provider-A", "--", "--help"},
			want: false,
		},
		{
			name: "empty args",
			args: []string{},
			want: false,
		},
		{
			name: "--help with provider",
			args: []string{"provider-A", "--help"},
			want: true,
		},
		{
			name: "provider only",
			args: []string{"provider-A"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WantsHelp(tt.args))
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantProvider   string
		wantTargetArgs []string
		wantUseTUI     bool
		wantErr        string
	}{
		{
			name:           "no args at all",
			args:           []string{},
			wantProvider:   "",
			wantTargetArgs: []string{},
			wantUseTUI:     true,
		},
		{
			name:           "single arg without separator",
			args:           []string{"mystage"},
			wantProvider:   "mystage",
			wantTargetArgs: []string{},
			wantUseTUI:     false,
		},
		{
			name:           "multiple args without separator",
			args:           []string{"prod", "--verbose", "--json"},
			wantProvider:   "prod",
			wantTargetArgs: []string{"--verbose", "--json"},
			wantUseTUI:     false,
		},
		{
			name:           "separator with one arg before",
			args:           []string{"dev", "--", "subcommand", "--flag"},
			wantProvider:   "dev",
			wantTargetArgs: []string{"subcommand", "--flag"},
			wantUseTUI:     false,
		},
		{
			name:           "separator with no arg before",
			args:           []string{"--", "subcommand"},
			wantProvider:   "",
			wantTargetArgs: []string{"subcommand"},
			wantUseTUI:     true,
		},
		{
			name:           "separator with multiple args before",
			args:           []string{"a", "b", "--", "cmd"},
			wantProvider:   "",
			wantTargetArgs: nil,
			wantUseTUI:     false,
			wantErr:        "at most one argument allowed before --",
		},
		{
			name:           "separator with nothing after",
			args:           []string{"staging", "--"},
			wantProvider:   "staging",
			wantTargetArgs: []string{},
			wantUseTUI:     false,
		},
		{
			name:           "only separator",
			args:           []string{"--"},
			wantProvider:   "",
			wantTargetArgs: []string{},
			wantUseTUI:     true,
		},
		{
			name:           "flag-like double dash in provider position",
			args:           []string{"--model", "foo"},
			wantProvider:   "",
			wantTargetArgs: nil,
			wantUseTUI:     false,
			wantErr:        "flag-like argument '--model'",
		},
		{
			name:           "flag-like single dash in provider position",
			args:           []string{"-m"},
			wantProvider:   "",
			wantTargetArgs: nil,
			wantUseTUI:     false,
			wantErr:        "flag-like argument '-m'",
		},
		{
			name:           "flag-like arg before separator",
			args:           []string{"--model", "--", "foo"},
			wantProvider:   "",
			wantTargetArgs: nil,
			wantUseTUI:     false,
			wantErr:        "flag-like argument '--model'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, targetArgs, useTUI, err := ParseArgs(tt.args)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantProvider, provider)
			assert.Equal(t, tt.wantTargetArgs, targetArgs)
			assert.Equal(t, tt.wantUseTUI, useTUI)
		})
	}
}
