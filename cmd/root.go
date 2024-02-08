package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var limit_flag int
var top_flag int
var language_flag string

func _root() error {

	rootCmd := &cobra.Command{
		Use:   "language <subcommand> [flags]",
		Short: "gh language",
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().IntVarP(&limit_flag, "limit", "l", 100, "The maximum number of repositories to evaluate")
	rootCmd.PersistentFlags().IntVarP(&top_flag, "top", "t", 10, "Return the top N languages (ignored when a language is specified)")
	rootCmd.PersistentFlags().StringVarP(&language_flag, "language", "L", "", "The language to filter on")

	rootCmd.MarkFlagsMutuallyExclusive("top", "language")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		return
	}

	rootCmd.AddCommand(countCmd)

	return rootCmd.Execute()
}

func Root() {
	if err := _root(); err != nil {
		fmt.Fprintf(os.Stderr, Red("Error: %s"), err.Error())
	}
}
