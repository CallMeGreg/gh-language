package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type languageDetails struct {
	Size int `json:"size"`
	Node struct {
		Name string `json:"name"`
	} `json:"node"`
}

type Repo struct {
	Languages     []languageDetails `json:"languages"`
	NameWithOwner string            `json:"nameWithOwner"`
}

type languageCount struct {
	name  string
	count int
}

func _root() error {

	rootCmd := &cobra.Command{
		Use:   "language <subcommand> [flags]",
		Short: "gh language",
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().IntP("limit", "l", 100, "The maximum number of repositories to evaluate")
	rootCmd.PersistentFlags().IntP("top", "t", 10, "Return the top N languages (ignored when a language is specified)")
	rootCmd.PersistentFlags().StringP("language", "L", "", "The language to filter on")

	// add mutually exclusive check for top and language

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		return
	}

	rootCmd.AddCommand(countCmd)

	return rootCmd.Execute()
}

func Root() {
	if err := _root(); err != nil {
		fmt.Fprintf(os.Stderr, "X %s", err.Error())
	}
}
