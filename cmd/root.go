package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var limit_flag int
var top_flag int
var language_flag string

var RootCmd = &cobra.Command{
	Use:   "language <subcommand> [flags]",
	Short: "gh language",
}

func _root() error {
	RootCmd.CompletionOptions.DisableDefaultCmd = true

	RootCmd.PersistentFlags().IntVarP(&limit_flag, "limit", "l", 100, "The maximum number of repositories to evaluate")
	RootCmd.PersistentFlags().IntVarP(&top_flag, "top", "t", 10, "Return the top N languages (ignored when a language is specified)")
	RootCmd.PersistentFlags().StringVarP(&language_flag, "language", "L", "", "The language to filter on")

	RootCmd.MarkFlagsMutuallyExclusive("top", "language")

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		return
	}

	RootCmd.AddCommand(countCmd)
	RootCmd.AddCommand(trendCmd)
	RootCmd.AddCommand(dataCmd)

	return RootCmd.Execute()
}

func Root() {
	if err := _root(); err != nil {
		fmt.Fprintf(os.Stderr, Red("Error: %s"), err.Error())
	}
}
