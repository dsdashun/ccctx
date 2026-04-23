package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/dsdashun/ccctx/config"
	"github.com/dsdashun/ccctx/internal/runner"
	"github.com/dsdashun/ccctx/internal/ui"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:                "run [context] [-- claude-args...]",
	Short:              "Run claude with a context",
	Long:               "Run claude with the specified context or interactively select one. Arguments after '--' are passed to claude.",
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if runner.WantsHelp(args) {
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			return
		}
		os.Exit(runRun(args))
	},
}

func runRun(args []string) int {
	model, smallFastModel, args, err := runner.ExtractFlags(args)
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

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: claude not found in PATH\n")
		return 1
	}

	target := append([]string{claudePath}, targetArgs...)

	r, err := runner.New(runner.Options{
		ContextName:    provider,
		Target:         target,
		Model:          model,
		SmallFastModel: smallFastModel,
	})
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
