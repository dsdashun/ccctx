package runner

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/dsdashun/ccctx/config"
)

type Options struct {
	ContextName    string
	Target         []string
	Model          string
	SmallFastModel string
	HaikuModel     string
	SonnetModel    string
	OpusModel      string
}

type Runner struct {
	ctx  *config.Context
	opts Options
	env  []string
}

func New(opts Options) (*Runner, error) {
	ctx, err := config.GetContext(opts.ContextName)
	if err != nil {
		return nil, err
	}
	if ctx.BaseURL == "" {
		return nil, fmt.Errorf("context '%s' is missing base_url", opts.ContextName)
	}
	if err := validateURL(ctx.BaseURL); err != nil {
		return nil, fmt.Errorf("context '%s': %w", opts.ContextName, err)
	}
	if ctx.AuthToken == "" {
		return nil, fmt.Errorf("context '%s' is missing auth_token", opts.ContextName)
	}
	if len(opts.Target) == 0 {
		return nil, fmt.Errorf("target command is required")
	}
	env := buildEnv(ctx, opts)
	return &Runner{ctx: ctx, opts: opts, env: env}, nil
}

func validateURL(rawURL string) error {
	if strings.Contains(rawURL, " ") {
		return fmt.Errorf("invalid base_url: contains spaces")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid base_url: %w", err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("invalid base_url: missing scheme (e.g., https://)")
	}
	return nil
}

// Run executes the target command. Returns (0, nil) on success, (exitCode, nil) for
// command exit errors, (1, error) for start failures. Caller is responsible for printing errors.
func (r *Runner) Run() (int, error) {
	cmd := exec.Command(r.opts.Target[0], r.opts.Target[1:]...)
	cmd.Env = r.env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

func buildEnv(ctx *config.Context, opts Options) []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "ANTHROPIC_") {
			filtered = append(filtered, e)
		}
	}
	filtered = append(filtered, "ANTHROPIC_BASE_URL="+ctx.BaseURL)
	filtered = append(filtered, "ANTHROPIC_AUTH_TOKEN="+ctx.AuthToken)

	// Model: opts > config > omit
	model := opts.Model
	if model == "" {
		model = ctx.Model
	}
	if model != "" {
		filtered = append(filtered, "ANTHROPIC_MODEL="+model)
	}

	// Haiku: opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel > omit
	haiku := opts.HaikuModel
	if haiku == "" {
		haiku = opts.SmallFastModel
	}
	if haiku == "" {
		haiku = ctx.HaikuModel
	}
	if haiku == "" {
		haiku = ctx.SmallFastModel
	}
	if haiku != "" {
		filtered = append(filtered, "ANTHROPIC_DEFAULT_HAIKU_MODEL="+haiku)
	}

	// Sonnet: opts > config > omit
	sonnet := opts.SonnetModel
	if sonnet == "" {
		sonnet = ctx.SonnetModel
	}
	if sonnet != "" {
		filtered = append(filtered, "ANTHROPIC_DEFAULT_SONNET_MODEL="+sonnet)
	}

	// Opus: opts > config > omit
	opus := opts.OpusModel
	if opus == "" {
		opus = ctx.OpusModel
	}
	if opus != "" {
		filtered = append(filtered, "ANTHROPIC_DEFAULT_OPUS_MODEL="+opus)
	}

	return filtered
}
