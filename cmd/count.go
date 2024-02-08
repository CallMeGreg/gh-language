package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cli/go-gh"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count [<org>]",
	Short: "Analyze the count of programming languages used in repos across an organization",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runCount(cmd, args)
	},
}

func runCount(cmd *cobra.Command, args []string) (err error) {

	var org string
	if len(args) > 0 {
		org = args[0]
	} else {
		return fmt.Errorf(Red("No organization specified."))
	}
	fmt.Printf(Yellow("Analyzing organization: %s\n"), org)

	language, _ := cmd.Flags().GetString("language")
	top, _ := cmd.Flags().GetInt("top")

	repoLimit, _ := cmd.Flags().GetInt("limit")
	if repoLimit > 0 {
		fmt.Printf(Yellow("Limiting to %d repositories.\n"), repoLimit)
	} else {
		fmt.Printf(Red("Please select a repository limit greater than 0.\n"))
		return
	}

	if language != "" {
		fmt.Printf(Yellow("Filtering on language: %s\n"), language)
	} else {
		if top > 0 {
			fmt.Printf(Yellow("Returning the top %d languages.\n"), top)
		} else {
			fmt.Printf(Red("Please select a top languages value greater than 0.\n"))
			return
		}
	}

	fmt.Println(Yellow("Fetching repositories and analyzing languages..."))

	repolanguages, _, err := gh.Exec("repo", "list", org, "--limit", strconv.FormatInt(int64(repoLimit), 10), "--json", "nameWithOwner,languages,createdAt")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	var repos []Repo
	jsonErr := json.Unmarshal(repolanguages.Bytes(), &repos)
	if jsonErr != nil {
		fmt.Printf("Error: %s\n", jsonErr.Error())
		return
	}

	// Sort repos by descending createdAt date
	sort.Slice(repos, func(i, j int) bool {
		iTime, _ := time.Parse(time.RFC3339, repos[i].CreatedAt)
		jTime, _ := time.Parse(time.RFC3339, repos[j].CreatedAt)
		return iTime.After(jTime)
	})

	fmt.Printf(Yellow("Analyzed %d repositories.\n"), len(repos))

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
	if language_flag != "" {
		// get the language details from the map:
		languageCount := languageMap[language_flag]
		fmt.Printf("%*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(analyzedRepoCount)), languageCount, 2, (languageCount*100)/analyzedRepoCount, language_flag)
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
