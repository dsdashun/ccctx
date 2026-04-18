package runner

import (
	"errors"
	"fmt"
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
	if ctx.AuthToken == "" {
		return nil, fmt.Errorf("context '%s' is missing auth_token", opts.ContextName)
	}
	if len(opts.Target) == 0 {
		return nil, fmt.Errorf("target command is required")
	}
	env := buildEnv(ctx, opts)
	return &Runner{ctx: ctx, opts: opts, env: env}, nil
}

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
	if ctx.Model != "" {
		filtered = append(filtered, "ANTHROPIC_MODEL="+ctx.Model)
	}
	if ctx.SmallFastModel != "" {
		filtered = append(filtered, "ANTHROPIC_SMALL_FAST_MODEL="+ctx.SmallFastModel)
	}
	return filtered
}
