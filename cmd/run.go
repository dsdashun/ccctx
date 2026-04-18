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

		var contextName string
		if useTUI {
			contexts, err := config.ListContexts()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(contexts) == 0 {
				fmt.Println("No contexts found.")
				return
			}

			selected, selectorErr := ui.RunContextSelector()
			if selectorErr != nil {
				if selectorErr.Error() == "operation cancelled" {
					fmt.Println("Operation cancelled.")
					return
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", selectorErr)
				os.Exit(1)
			}
			contextName = selected
		} else {
			contextName = provider
		}

		claudePath, err := exec.LookPath("claude")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: claude not found in PATH\n")
			os.Exit(1)
		}

		target := append([]string{claudePath}, targetArgs...)

		r, err := runner.New(runner.Options{
			ContextName: contextName,
			Target:      target,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		exitCode, err := r.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing claude: %v\n", err)
			os.Exit(1)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	},
}
