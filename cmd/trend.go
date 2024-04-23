package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

const GITHUB_TIMESTAMP_LAYOUT = "2006-01-02T15:04:05Z"

var trendCmd = &cobra.Command{
	Use:   "trend [<org>]",
	Short: "Analyze the trend of programming languages used in repos across an organization over time",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runTrend(cmd, args)
	},
}

func runTrend(cmd *cobra.Command, args []string) (err error) {
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

	languageMapPerYear := make(map[int]map[string]int)

	for _, repo := range repos {
		t, err := time.Parse(GITHUB_TIMESTAMP_LAYOUT, repo.CreatedAt)
		if err != nil {
			fmt.Print(Red("Error parsing CreatedAt date. Skipping repo: " + repo.NameWithOwner + "\n"))
			continue
		}
		year := t.Year()
		if _, ok := languageMapPerYear[year]; !ok {
			languageMapPerYear[year] = make(map[string]int)
		}
		for _, language := range repo.Languages {
			languageMapPerYear[year][language.Node.Name]++
		}
	}

	for year, languageMap := range languageMapPerYear {
		// convert the map to a slice
		languages := make([]languageCount, 0, len(languageMap))
		for name, count := range languageMap {
			languages = append(languages, languageCount{name, count})
		}

		// sort the slice by count in descending order
		sort.Slice(languages, func(i, j int) bool {
			return languages[i].count > languages[j].count
		})
		if language_flag == "" {
			fmt.Printf(Green("Repos created in year: %d\n"), year)
			// print out the top N languages, along with the percent frequency
			counter := 0
			for _, lang := range languages {
				if counter >= top {
					break
				}
				fmt.Printf("  %*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(len(repos))), lang.count, 2, (lang.count*100)/len(repos), lang.name)
				counter++
			}
			// if a language was specified, only print that result:
		} else {
			for _, lang := range languages {
				if lang.name == language_flag {
					fmt.Printf(Green("Repos created in year: %d\n"), year)
					fmt.Printf("  %*d repos (%*d%%) that include the language: %s\n", len(strconv.Itoa(len(repos))), lang.count, 2, (lang.count*100)/len(repos), lang.name)
					break
				}
			}
		}
	}

	return
}
