package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var enterprise_flag string
var org_flag string
var org_limit_flag int
var repo_limit_flag int
var top_flag int
var language_flag string
var codeql_flag bool

var RootCmd = &cobra.Command{
	Use:   "language <subcommand> [flags]",
	Short: "gh language",
}

func _root() error {
	RootCmd.CompletionOptions.DisableDefaultCmd = true

	RootCmd.PersistentFlags().StringVarP(&enterprise_flag, "enterprise", "e", "", "Specify the enterprise")
	RootCmd.PersistentFlags().StringVarP(&org_flag, "org", "o", "", "Specify the organization")
	RootCmd.PersistentFlags().IntVar(&org_limit_flag, "org-limit", 5, "The maximum number of organizations to analyze for an enterprise")
	RootCmd.PersistentFlags().IntVar(&repo_limit_flag, "repo-limit", 10, "The maximum number of repositories to analyze per organization")
	RootCmd.PersistentFlags().IntVarP(&top_flag, "top", "t", 10, "Return the top N languages (ignored when a language filter is specified)")
	RootCmd.PersistentFlags().StringVarP(&language_flag, "language", "l", "", "A specific language to filter on (case-sensitive)")
	RootCmd.PersistentFlags().BoolVar(&codeql_flag, "codeql", false, "Restrict analysis to CodeQL-supported languages")

	RootCmd.MarkFlagsMutuallyExclusive("enterprise", "org")
	RootCmd.MarkFlagsMutuallyExclusive("top", "language", "codeql")

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
