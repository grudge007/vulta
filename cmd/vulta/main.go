package main

import (
	"fmt"
	"vulta/cmd/commands"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(commands.InitCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return
	}
}

var rootCmd = &cobra.Command{
	Use:           "vulta",
	Short:         "A lightweight deployment manager",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
