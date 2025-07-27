package main

import (
	"fmt"
	"os"

	"github.com/dsdashun/ccctx/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ccctx",
	Short: "Claude Context Switcher",
	Long:  "A CLI tool to manage and switch between different Claude contexts",
}

func init() {
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.SwitchCmd)
	rootCmd.AddCommand(cmd.RunCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}