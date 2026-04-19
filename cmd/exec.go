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
	Use:   "exec [context] [-- command...]",
	Short: "Execute a command or launch a shell with a context",
	Long:  "Execute a command or launch a shell with the specified context. If no command is given, launches $SHELL. If no context is given, opens the interactive selector.",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		provider, targetArgs, useTUI, err := runner.ParseArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if useTUI {
			contexts, err := config.ListContexts()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if len(contexts) == 0 {
				fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
				os.Exit(1)
			}
			provider, err = ui.RunContextSelector(contexts)
			if err != nil {
				if errors.Is(err, ui.ErrCancelled) {
					fmt.Fprintln(os.Stderr, "Operation cancelled.")
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		if len(targetArgs) == 0 {
			shell := os.Getenv("SHELL")
			if shell == "" {
				fmt.Fprintf(os.Stderr, "Error: SHELL environment variable not set\n")
				os.Exit(1)
			}
			targetArgs = []string{shell}
		}

		opts := runner.Options{ContextName: provider, Target: targetArgs}
		r, err := runner.New(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		exitCode, err := r.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(exitCode)
	},
}
