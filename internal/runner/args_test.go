package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantModel     string
		wantSmallFast string
		wantRemaining []string
		wantErr       string
	}{
		{
			name:          "no flags returns empty strings and args unchanged",
			args:          []string{"provider-A"},
			wantModel:     "",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "--model before provider",
			args:          []string{"--model", "foo", "provider-A"},
			wantModel:     "foo",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "--model after provider",
			args:          []string{"provider-A", "--model", "foo"},
			wantModel:     "foo",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "--model with separator preserves forwarded args",
			args:          []string{"provider-A", "--model", "foo", "--", "--help"},
			wantModel:     "foo",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A", "--", "--help"},
		},
		{
			name:          "both flags extracted",
			args:          []string{"provider-A", "--model", "foo", "--small-fast-model", "bar", "--", "cmd"},
			wantModel:     "foo",
			wantSmallFast: "bar",
			wantRemaining: []string{"provider-A", "--", "cmd"},
		},
		{
			name:          "--model without value returns error",
			args:          []string{"--model"},
			wantErr:       "--model requires a value",
			wantModel:     "",
			wantSmallFast: "",
			wantRemaining: []string{},
		},
		{
			name:          "--small-fast-model without value returns error",
			args:          []string{"--small-fast-model"},
			wantErr:       "--small-fast-model requires a value",
			wantModel:     "",
			wantSmallFast: "",
			wantRemaining: []string{},
		},
		{
			name:          "--model after separator is NOT extracted",
			args:          []string{"provider-A", "--", "--model", "foo"},
			wantModel:     "",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A", "--", "--model", "foo"},
		},
		{
			name:          "no provider with flag",
			args:          []string{"--model", "foo"},
			wantModel:     "foo",
			wantSmallFast: "",
			wantRemaining: []string{},
		},
		{
			name:          "flag value is dash",
			args:          []string{"provider-A", "--model", "-"},
			wantModel:     "-",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "duplicate flags last wins",
			args:          []string{"provider-A", "--model", "foo", "--model", "bar"},
			wantModel:     "bar",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "duplicate --small-fast-model last wins",
			args:          []string{"provider-A", "--small-fast-model", "foo", "--small-fast-model", "bar"},
			wantModel:     "",
			wantSmallFast: "bar",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "flag value looks like another flag",
			args:          []string{"provider-A", "--model", "--small-fast-model"},
			wantModel:     "--small-fast-model",
			wantSmallFast: "",
			wantRemaining: []string{"provider-A"},
		},
		{
			name:          "separator with no following args",
			args:          []string{"--model", "foo", "--"},
			wantModel:     "foo",
			wantSmallFast: "",
			wantRemaining: []string{"--"},
		},
		{
			name:          "--model value with newline returns error",
			args:          []string{"provider-A", "--model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantSmallFast: "",
			wantRemaining: []string{},
		},
		{
			name:          "--small-fast-model value with newline returns error",
			args:          []string{"provider-A", "--small-fast-model", "foo\nbar"},
			wantErr:       "value cannot contain newline",
			wantModel:     "",
			wantSmallFast: "",
			wantRemaining: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, sfm, remaining, err := ExtractFlags(tt.args)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Equal(t, []string{}, remaining)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantModel, model)
			assert.Equal(t, tt.wantSmallFast, sfm)
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
