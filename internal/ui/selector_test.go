package ui

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrCancelled(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    bool
		message string
	}{
		{
			name:    "ErrCancelled matches via errors.Is",
			err:     ErrCancelled,
			want:    true,
			message: "operation cancelled",
		},
		{
			name:    "wrapped ErrCancelled matches via errors.Is",
			err:     fmt.Errorf("wrapper: %w", ErrCancelled),
			want:    true,
			message: "operation cancelled",
		},
		{
			name:    "other error does not match ErrCancelled",
			err:     errors.New("something else"),
			want:    false,
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, ErrCancelled)
			assert.Equal(t, tt.want, result)
			if tt.message != "" {
				assert.Contains(t, tt.err.Error(), tt.message)
			}
		})
	}
}

func TestRunContextSelector_EmptyContexts(t *testing.T) {
	_, err := RunContextSelector([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no contexts found")
}
