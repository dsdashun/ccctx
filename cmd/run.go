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
	Use:   "run [context] [-- claude-args...]",
	Short: "Run claude with a context",
	Long:  "Run claude with the specified context or interactively select one. Arguments after '--' are passed to claude.",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var contextName string
		var claudeArgs []string
		
		// Find the position of '--' separator if it exists
		separatorIndex := -1
		for i, arg := range args {
			if arg == "--" {
				separatorIndex = i
				break
			}
		}
		
		// Parse arguments
		if separatorIndex != -1 {
			// Everything before '--' are our args, everything after are claude args
			contextArgs := args[:separatorIndex]
			claudeArgs = args[separatorIndex+1:]
			
			if len(contextArgs) > 0 {
				contextName = contextArgs[0]
			} else {
				// Interactive mode - use the UI selector
				contexts, err := config.ListContexts()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				if len(contexts) == 0 {
					fmt.Println("No contexts found.")
					return
				}
				
				var selectorErr error
				contextName, selectorErr = ui.RunContextSelector()
				if selectorErr != nil {
					if selectorErr.Error() == "operation cancelled" {
						fmt.Println("Operation cancelled.")
						return
					}
					fmt.Fprintf(os.Stderr, "Error: %v\n", selectorErr)
					os.Exit(1)
				}
			}
		} else {
			// No separator, check if first arg is a context name
			if len(args) > 0 {
				contexts, err := config.ListContexts()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				
				// If first arg matches a context, use it as context name
				isContext := false
				for _, ctx := range contexts {
					if args[0] == ctx {
						isContext = true
						break
					}
				}
				
				if isContext {
					contextName = args[0]
					// All other args are claude args
					claudeArgs = args[1:]
				} else {
					// First arg is not a context, so we're in interactive mode
					// All args are claude args
					claudeArgs = args
					
					// Interactive mode - use the UI selector
					contexts, err := config.ListContexts()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						os.Exit(1)
					}

					if len(contexts) == 0 {
						fmt.Println("No contexts found.")
						return
					}
					
					var selectorErr error
					contextName, selectorErr = ui.RunContextSelector()
					if selectorErr != nil {
						if selectorErr.Error() == "operation cancelled" {
							fmt.Println("Operation cancelled.")
							return
						}
						fmt.Fprintf(os.Stderr, "Error: %v\n", selectorErr)
						os.Exit(1)
					}
				}
			} else {
				// Interactive mode - use the UI selector
				contexts, err := config.ListContexts()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				if len(contexts) == 0 {
					fmt.Println("No contexts found.")
					return
				}
				
				var selectorErr error
				contextName, selectorErr = ui.RunContextSelector()
				if selectorErr != nil {
					if selectorErr.Error() == "operation cancelled" {
						fmt.Println("Operation cancelled.")
						return
					}
					fmt.Fprintf(os.Stderr, "Error: %v\n", selectorErr)
					os.Exit(1)
				}
			}
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

		// Execute claude with the modified environment and forwarded args
		cmdClaude := exec.Command(claudePath, claudeArgs...)
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