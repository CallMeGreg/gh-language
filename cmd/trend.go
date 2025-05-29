package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/cli/go-gh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const GITHUB_TIMESTAMP_LAYOUT = "2006-01-02T15:04:05Z"

var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Analyze the trend of programming languages used in repos across an enterprise or organization over time",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runTrend(cmd, args)
	},
}

func runTrend(cmd *cobra.Command, args []string) error {
	org := org_flag
	enterprise := enterprise_flag
	repoLimit := repo_limit_flag
	orgLimit := org_limit_flag
	top := top_flag
	language := language_flag

	if err := ValidateFlags(org, enterprise); err != nil {
		return err
	}

	var orgs []string

	if enterprise != "" {
		// Determine the language filter or top languages info based on flags.
		languageFilter := GetLanguageFilter(codeql_flag, language, top)
		// Print organization and repository limits along with the language filter.
		PrintInfoWithFormat("Organization limit: %d, Repository limit: %d, %s", orgLimit, repoLimit, languageFilter)
		PrintIndexingEnterprise(enterprise)
		var err error
		orgs, err = FetchOrganizations(enterprise, orgLimit)
		if err != nil {
			return err
		}
	} else {
		// Handle the case where only a single organization is provided.
		topLanguagesInfo := GetLanguageFilter(codeql_flag, language, top)
		PrintInfoWithFormat("Repository limit: %d, %s", repoLimit, topLanguagesInfo)
		orgs = []string{org}
	}

	// Initialize a map to store language data per year.
	languageMapPerYear := make(map[int]map[string]int)

	var totalRepos int

	// Iterate over each organization to fetch repositories and analyze languages.
	for _, org := range orgs {
		// Start a spinner to indicate progress for indexing the organization.
		spinnerInfo, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Indexing organization: %s", org))

		// Fetch repositories for the organization. This involves a REST API call to GitHub.
		repos, err := FetchRepositories(org, repoLimit)
		if err != nil {
			// Stop the spinner and indicate failure if an error occurs.
			spinnerInfo.Fail("Failed to index organization")
			return err
		}

		if len(repos) == 0 {
			// Stop the spinner and indicate a warning if no repositories are found.
			spinnerInfo.Warning(fmt.Sprintf("No repositories found for organization: %s", org))
			continue
		}

		// Stop the spinner and indicate success.
		spinnerInfo.Success(fmt.Sprintf("Successfully indexed organization: %s", org))
		// Start a progress bar for analyzing repositories.
		progressBar, _ := pterm.DefaultProgressbar.WithTotal(len(repos)).WithTitle("Analyzing repositories").Start()

		// Increment the total repository count.
		totalRepos += len(repos)

		// Analyze each repository for language usage and group by year.
		for _, repo := range repos {
			progressBar.Increment()
			// Fetch language data for the repository. This involves another REST API call to GitHub.
			output, _, err := gh.Exec("api", fmt.Sprintf("repos/%s/%s/languages", org, repo.Name))
			if err != nil {
				// Print a warning and skip the repository if an error occurs.
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to error: %s", repo.Name, err))
				continue
			}

			// Parse the language data from the API response. This step can fail if the response format changes.
			var repoLanguages map[string]int
			if err := json.Unmarshal(output.Bytes(), &repoLanguages); err != nil {
				// Print a warning and skip the repository if parsing fails.
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to parsing error: %s", repo.Name, err))
				continue
			}

			// Parse the repository's creation date
			createdAt, err := time.Parse(GITHUB_TIMESTAMP_LAYOUT, repo.CreatedAt)
			if err != nil {
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to invalid creation date: %s", repo.Name, err))
				continue
			}

			// Group the language data by year. This requires extracting the year from the repository's creation date.
			creationYear := createdAt.Year()
			if languageMapPerYear[creationYear] == nil {
				languageMapPerYear[creationYear] = make(map[string]int)
			}

			for lang := range repoLanguages {
				languageMapPerYear[creationYear][lang]++
			}
		}

		// Stop the progress bar after analyzing all repositories.
		progressBar.Stop()
	}

	// Print the total number of repositories analyzed.
	pterm.Println() // Add a new line
	pterm.Info.Println(fmt.Sprintf("Total number of repositories analyzed: %d", totalRepos))
	pterm.Println() // Add a new line

	// Extract the keys (years) into a slice for sorting.
	years := make([]int, 0, len(languageMapPerYear))
	for year := range languageMapPerYear {
		years = append(years, year)
	}

	// Sort the slice in ascending order.
	sort.Ints(years)

	// Reverse the slice to get it in descending order
	for i, j := 0, len(years)-1; i < j; i, j = i+1, j-1 {
		years[i], years[j] = years[j], years[i]
	}

	if codeql_flag {
		for year, langMap := range languageMapPerYear {
			languageMapPerYear[year] = IsCodeQLLanguage(langMap)
		}
	}

	// Update percentage calculation
	for _, year := range years {
		pterm.DefaultSection.Println(fmt.Sprintf("Year: %d", year))
		sortedLanguages := make([]struct {
			Language string
			Count    int
		}, 0, len(languageMapPerYear[year]))

		for lang, count := range languageMapPerYear[year] {
			if language != "" && lang != language {
				continue
			}
			sortedLanguages = append(sortedLanguages, struct {
				Language string
				Count    int
			}{lang, count})
		}

		sort.Slice(sortedLanguages, func(i, j int) bool {
			return sortedLanguages[i].Count > sortedLanguages[j].Count
		})

		rows := [][]string{{"Language", "Count", "Percentage"}}
		for i, langData := range sortedLanguages {
			if top > 0 && i >= top {
				break
			}
			percentage := int(float64(langData.Count) / float64(totalRepos) * 100)
			rows = append(rows, []string{langData.Language, fmt.Sprintf("%d", langData.Count), fmt.Sprintf("%d%%", percentage)})
		}

		pterm.DefaultTable.WithHasHeader(true).WithData(rows).Render()
	}

	return nil
}
