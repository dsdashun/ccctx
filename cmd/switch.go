package cmd

import (
	"fmt"
	"os"

	"github.com/dsdashun/ccctx/config"
	"github.com/manifoldco/promptui"
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

			// Create interactive selector
			prompt := promptui.Select{
				Label: "Select a context",
				Items: contexts,
				Size:  10, // Show 10 items at a time
			}

			_, result, err := prompt.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			printExportCommands(result)
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