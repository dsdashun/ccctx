package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dsdashun/ccctx/config"
	"github.com/dsdashun/ccctx/internal/runner"
	"github.com/dsdashun/ccctx/internal/ui"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [context] [-- claude-args...]",
	Short: "Run claude with a context",
	Long:  "Run claude with the specified context or interactively select one. Arguments after '--' are passed to claude.",
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

			provider, err = ui.RunContextSelector()
			if err != nil {
				if err.Error() == "operation cancelled" {
					fmt.Println("Operation cancelled.")
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		claudePath, err := exec.LookPath("claude")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: claude not found in PATH\n")
			os.Exit(1)
		}

		target := append([]string{claudePath}, targetArgs...)

		r, err := runner.New(runner.Options{
			ContextName: provider,
			Target:      target,
		})
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
