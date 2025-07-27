package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dsdashun/ccctx/config"
	"github.com/dsdashun/ccctx/internal/ui"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [context]",
	Short: "Run claude with a context",
	Long:  "Run claude with the specified context or interactively select one",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var contextName string
		
		if len(args) == 0 {
			// Interactive mode - use the UI selector
			var err error
			contextName, err = ui.RunContextSelector()
			if err != nil {
				if err.Error() == "operation cancelled" {
					fmt.Println("Operation cancelled.")
					return
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			contextName = args[0]
		}

		// Get context details
		ctx, err := config.GetContext(contextName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Find claude binary
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: claude not found in PATH\n")
			os.Exit(1)
		}

		// Prepare environment with context variables
		// Start with a clean environment based on the current one
		env := os.Environ()
		
		// Remove any existing ANTHROPIC_BASE_URL and ANTHROPIC_AUTH_TOKEN
		newEnv := []string{}
		for _, e := range env {
			if !(len(e) >= 19 && e[:19] == "ANTHROPIC_BASE_URL=") && 
			   !(len(e) >= 20 && e[:20] == "ANTHROPIC_AUTH_TOKEN=") {
				newEnv = append(newEnv, e)
			}
		}
		
		// Add the new context variables
		newEnv = append(newEnv, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", ctx.BaseURL))
		newEnv = append(newEnv, fmt.Sprintf("ANTHROPIC_AUTH_TOKEN=%s", ctx.AuthToken))

		// Execute claude with the modified environment
		cmdClaude := exec.Command(claudePath)
		cmdClaude.Env = newEnv
		cmdClaude.Stdin = os.Stdin
		cmdClaude.Stdout = os.Stdout
		cmdClaude.Stderr = os.Stderr
		
		// Run claude and wait for it to complete
		if err := cmdClaude.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// Propagate the exit code from claude
				os.Exit(exitError.ExitCode())
			} else {
				fmt.Fprintf(os.Stderr, "Error executing claude: %v\n", err)
				os.Exit(1)
			}
		}
	},
}