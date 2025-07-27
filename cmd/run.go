package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/dsdashun/ccctx/config"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [context]",
	Short: "Run claude-code with a context",
	Long:  "Run claude-code with the specified context or interactively select one",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var contextName string
		
		if len(args) == 0 {
			// Interactive mode
			contexts, err := config.ListContexts()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(contexts) == 0 {
				fmt.Println("No contexts found.")
				return
			}

			fmt.Println("Available contexts:")
			for i, ctx := range contexts {
				fmt.Printf("%d. %s\n", i+1, ctx)
			}

			fmt.Print("Enter context number: ")
			var choice int
			_, err = fmt.Scanf("%d", &choice)
			if err != nil || choice < 1 || choice > len(contexts) {
				fmt.Fprintf(os.Stderr, "Invalid choice\n")
				os.Exit(1)
			}

			contextName = contexts[choice-1]
		} else {
			contextName = args[0]
		}

		// Get context details
		ctx, err := config.GetContext(contextName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Find claude-code binary
		claudeCodePath, err := exec.LookPath("claude-code")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: claude-code not found in PATH\n")
			os.Exit(1)
		}

		// Prepare environment with context variables
		env := os.Environ()
		env = append(env, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", ctx.BaseURL))
		env = append(env, fmt.Sprintf("ANTHROPIC_AUTH_TOKEN=%s", ctx.AuthToken))

		// Execute claude-code with the modified environment
		binary, err := exec.LookPath(claudeCodePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding claude-code: %v\n", err)
			os.Exit(1)
		}

		// Execute claude-code with the specified context
		err = syscall.Exec(binary, []string{"claude-code"}, env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing claude-code: %v\n", err)
			os.Exit(1)
		}
	},
}