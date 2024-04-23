package cmd

import (
	"fmt"
	"sort"
	"strconv"

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
		fmt.Print(Red("Please select a repository limit greater than 0.\n"))
		return
	}

	if language != "" {
		fmt.Printf(Yellow("Filtering on language: %s\n"), language)
	} else {
		if top > 0 {
			fmt.Printf(Yellow("Returning the top %d languages.\n"), top)
		} else {
			fmt.Print(Red("Please select a top languages value greater than 0.\n"))
			return
		}
	}

	repos, err := getAllRepos(org, language, repoLimit)

	languageMap := make(map[string]int)

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
		fmt.Printf("%*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(len(repos))), languageCount, 2, (languageCount*100)/len(repos), language_flag)
	} else {
		// print out the top N languages, along with the percent frequency
		counter := 0
		for _, lang := range languages {
			if counter >= top {
				break
			}
			fmt.Printf("%*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(len(repos))), lang.count, 2, (lang.count*100)/len(repos), lang.name)
			counter++
		}
	}

	return
}
