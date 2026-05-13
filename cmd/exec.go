package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/dsdashun/ccctx/config"
	"github.com/dsdashun/ccctx/internal/runner"
	"github.com/dsdashun/ccctx/internal/ui"
	"github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
	Use:                "exec [context] [-- command...]",
	Short:              "Execute a command or launch a shell with a context",
	Long:               "Execute a command or launch a shell with the specified context. If no command is given, launches $SHELL. If no context is given, opens the interactive selector.",
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if runner.WantsHelp(args) {
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			return
		}
		os.Exit(execRun(args))
	},
}

func execRun(args []string) int {
	model, haikuModel, sonnetModel, opusModel, args, err := runner.ExtractFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	provider, targetArgs, useTUI, err := runner.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if useTUI {
		contexts, err := config.ListContexts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if len(contexts) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
			return 1
		}
		provider, err = ui.RunContextSelector(contexts)
		if err != nil {
			if errors.Is(err, ui.ErrCancelled) {
				fmt.Fprintln(os.Stderr, "Operation cancelled.")
				return 1
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	if len(targetArgs) == 0 {
		shell := os.Getenv("SHELL")
		if shell == "" {
			fmt.Fprintf(os.Stderr, "Error: SHELL environment variable not set\n")
			return 1
		}
		targetArgs = []string{shell}
	}

	opts := runner.Options{ContextName: provider, Target: targetArgs, Model: model, HaikuModel: haikuModel, SonnetModel: sonnetModel, OpusModel: opusModel}
	r, err := runner.New(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	exitCode, err := r.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return exitCode
}
