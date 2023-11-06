package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

func _main() error {
	rootCmd := &cobra.Command{
		Use:   "language <subcommand> [flags]",
		Short: "gh language",
	}

	rootCmd.PersistentFlags().IntP("limit", "l", 100, "The maximum number of repositories to evaluate")
	rootCmd.PersistentFlags().IntP("top", "t", 10, "Return the top N languages")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		return
	}

	countCmd := &cobra.Command{
		Use:   "count [<org>]",
		Short: "Analyze the count of programming languages used in repos across an organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return runCount(cmd, args)
		},
	}
	rootCmd.AddCommand(countCmd)

	return rootCmd.Execute()
}

func runCount(cmd *cobra.Command, args []string) (err error) {

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

	repoLimit, _ := cmd.Flags().GetInt("limit")
	if repoLimit > 0 {
		fmt.Printf("Limiting to %d repositories.\n", repoLimit)
	}

	top, _ := cmd.Flags().GetInt("top")
	if top > 0 {
		fmt.Printf("Returning the top %d languages.\n", top)
	}

	repo_cmd := exec.Command("gh", "repo", "list", org, "--limit", strconv.FormatInt(int64(repoLimit), 10), "--json", "nameWithOwner,languages")
	repolanguages, err := repo_cmd.Output()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	type jsonLanguage struct {
		Size int `json:"size"`
		Node struct {
			Name string `json:"name"`
		} `json:"node"`
	}

	type Repo struct {
		Languages []jsonLanguage `json:"languages"`
	}

	var repos []Repo

	jsonErr := json.Unmarshal(repolanguages, &repos)
	if jsonErr != nil {
		fmt.Printf("Error: %s\n", jsonErr.Error())
		return
	}

	languageCount := make(map[string]int)
	repoCount := len(repos)

	for _, repo := range repos {
		for _, language := range repo.Languages {
			languageCount[language.Node.Name]++
		}
	}

	type outputLanguage struct {
		name  string
		count int
	}

	// convert the map to a slice
	languages := make([]outputLanguage, 0, len(languageCount))
	for name, count := range languageCount {
		languages = append(languages, outputLanguage{name, count})
	}

	// sort the slice by count in descending order
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].count > languages[j].count
	})

	// print out the top N languages, along with the percent frequency
	counter := 0
	for _, lang := range languages {
		if counter >= top {
			break
		}
		fmt.Printf("%*d repos (%d%%) that include the language: %s\n", len(strconv.Itoa(repoCount)), lang.count, (lang.count*100)/repoCount, lang.name)
		counter++
	}

	return
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "X %s", err.Error())
	}
}
