package cmd

import (
	"fmt"
	"os"

	"github.com/dsdashun/ccctx/config"
	"github.com/spf13/cobra"
)

var SwitchCmd = &cobra.Command{
	Use:   "switch [context]",
	Short: "Switch to a context",
	Long:  "Switch to a specified context or interactively select one",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

			selectedContext := contexts[choice-1]
			printExportCommands(selectedContext)
		} else {
			// Direct mode
			printExportCommands(args[0])
		}
	},
}

func printExportCommands(contextName string) {
	ctx, err := config.GetContext(contextName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("export ANTHROPIC_BASE_URL=%s\n", ctx.BaseURL)
	fmt.Printf("export ANTHROPIC_AUTH_TOKEN=%s\n", ctx.AuthToken)
}