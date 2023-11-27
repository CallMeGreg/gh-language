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

func _main() error {

	rootCmd := &cobra.Command{
		Use:   "language <subcommand> [flags]",
		Short: "gh language",
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().IntP("limit", "l", 100, "The maximum number of repositories to evaluate")
	rootCmd.PersistentFlags().IntP("top", "t", 10, "Return the top N languages (ignored when a language is specified)")
	rootCmd.PersistentFlags().StringP("language", "L", "", "The language to filter on")

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

	languageFilter, _ := cmd.Flags().GetString("language")
	top, _ := cmd.Flags().GetInt("top")

	repoLimit, _ := cmd.Flags().GetInt("limit")
	if repoLimit > 0 {
		fmt.Printf("Limiting to %d repositories.\n", repoLimit)
	} else {
		fmt.Printf("Please select a repository limit greater than 0.\n")
		return
	}

	if languageFilter != "" {
		fmt.Printf("Filtering on language: %s\n", languageFilter)
	} else {
		if top > 0 {
			fmt.Printf("Returning the top %d languages.\n", top)
		} else {
			fmt.Printf("Please select a top languages value greater than 0.\n")
			return
		}
	}

	repo_cmd := exec.Command("gh", "repo", "list", org, "--limit", strconv.FormatInt(int64(repoLimit), 10), "--json", "nameWithOwner,languages")
	repolanguages, err := repo_cmd.Output()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	var repos []Repo

	jsonErr := json.Unmarshal(repolanguages, &repos)
	if jsonErr != nil {
		fmt.Printf("Error: %s\n", jsonErr.Error())
		return
	}

	fmt.Printf("Identified %d repositories. Beginning analysis...\n", len(repos))

	languageMap := make(map[string]int)
	analyzedRepoCount := len(repos)

	for _, repo := range repos {
		for _, language := range repo.Languages {
			languageMap[language.Node.Name]++
		}
	}

	// convert the map to a slice
	languages := make([]languageCount, 0, len(languageMap))
	for name, count := range languageMap {
		languages = append(languages, languageCount{name, count})
	}

	// sort the slice by count in descending order
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].count > languages[j].count
	})

	// if a language was specified, only print that result:
	if languageFilter != "" {
		// get the language details from the map:
		languageCount := languageMap[languageFilter]
		fmt.Printf("%*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(analyzedRepoCount)), languageCount, 2, (languageCount*100)/analyzedRepoCount, languageFilter)
	} else {
		// print out the top N languages, along with the percent frequency
		counter := 0
		for _, lang := range languages {
			if counter >= top {
				break
			}
			fmt.Printf("%*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(analyzedRepoCount)), lang.count, 2, (lang.count*100)/analyzedRepoCount, lang.name)
			counter++
		}
	}

	return
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "X %s", err.Error())
	}
}
