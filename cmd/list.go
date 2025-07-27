package cmd

import (
	"fmt"
	"os"

	"github.com/dsdashun/ccctx/config"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available contexts",
	Long:  "List all available contexts from the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
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
		for _, ctx := range contexts {
			fmt.Printf("  %s\n", ctx)
		}
	},
}