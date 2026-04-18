package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
