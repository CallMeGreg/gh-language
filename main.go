package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func _main() error {
	rootCmd := &cobra.Command{
		Use:   "language <subcommand> [flags]",
		Short: "gh language",
	}

	rootCmd.PersistentFlags().BoolP("primary", "P", false, "Boolean value to only consider the primary language of each reposiotry")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		return
	}

	distributionCmd := &cobra.Command{
		Use:   "distribution [<org>]",
		Short: "Analyze the distribution of programming languages used in repos across an organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return runDistribution(cmd, args)
		},
	}

	dataCmd := &cobra.Command{
		Use:   "data [<org>]",
		Short: "Analyze the bytes of code written per programming language across an organization",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// ...
			return
		},
	}

	rootCmd.AddCommand(distributionCmd)
	rootCmd.AddCommand(dataCmd)

	return rootCmd.Execute()
}

func runDistribution(cmd *cobra.Command, args []string) (err error) {
	var org string
	if len(args) > 0 {
		org = args[0]
	} else {
		org = os.Getenv("GITHUB_ORG")
		if org == "" {
			return fmt.Errorf("No organization specified.")
		}
	}
	fmt.Printf("Analyzing organization: %s\n", org)

	primary, _ := cmd.Flags().GetBool("primary")
	if primary {
		fmt.Println("Only considering primary languages.")
	}

	return
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "X %s", err.Error())
	}
}
