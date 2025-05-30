package cmd

import (
	"fmt"
	"sort"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Analyze the count of programming languages used in repos across an enterprise or organization",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runCount(cmd, args)
	},
}

func runCount(cmd *cobra.Command, args []string) error {
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

	// Initialize a map to store language data and a counter for total repositories.
	languageData := make(map[string]int)
	var totalRepos int

	// Create the REST client once.
	client, err := api.DefaultRESTClient()
	if err != nil {
		pterm.Error.Println("Failed to create REST client:", err)
		return err
	}

	// Iterate over each organization to fetch repositories and analyze languages.
	for _, org := range orgs {
		// Start a spinner to indicate progress for indexing the organization.
		spinnerInfo, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Indexing organization: %s", org))

		// Fetch repositories for the organization. This involves a REST API call to GitHub.
		repos, err := FetchRepositories(client, org, repoLimit)
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

		// Analyze each repository for language usage.
		for _, repo := range repos {
			progressBar.Increment()
			// Fetch language data for the repository using FetchLanguages.
			languages, err := FetchLanguages(client, org, repo.Name)
			if err != nil {
				// Print a warning and skip the repository if an error occurs.
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to error: %s", repo.Name, err))
				continue
			}
			// Update the language data map with the fetched data by incrementing the count.
			for lang := range languages {
				languageData[lang]++
			}
		}

		// Stop the progress bar after analyzing all repositories.
		progressBar.Stop()
	}

	// Print the total number of repositories analyzed.
	pterm.Println() // Add a new line
	pterm.Info.Println(fmt.Sprintf("Total number of repositories analyzed: %d", totalRepos))
	pterm.Println() // Add a new line

	// Filter language data if a specific language is specified.
	if language != "" {
		// Create a new map to store only the filtered language data.
		filteredLanguageData := make(map[string]int)
		for lang, count := range languageData {
			if lang == language {
				filteredLanguageData[lang] = count
			}
		}
		languageData = filteredLanguageData
	}

	// Respect the --top flag by limiting the number of languages displayed.
	if top > 0 {
		// Sort the languages by their usage count in descending order.
		sortedLanguages := make([]struct {
			Language string
			Count    int
		}, 0, len(languageData))

		for lang, count := range languageData {
			sortedLanguages = append(sortedLanguages, struct {
				Language string
				Count    int
			}{lang, count})
		}

		sort.Slice(sortedLanguages, func(i, j int) bool {
			return sortedLanguages[i].Count > sortedLanguages[j].Count
		})

		// Select the top N languages based on the --top flag.
		topLanguages := make(map[string]int)
		for i := 0; i < top && i < len(sortedLanguages); i++ {
			topLanguages[sortedLanguages[i].Language] = sortedLanguages[i].Count
		}

		languageData = topLanguages
	}

	// Filter language data to include only CodeQL-supported languages if the flag is set.
	if codeql_flag {
		languageData = IsCodeQLLanguage(languageData)
	}

	// Render the language data as a table with percentages.
	pterm.DefaultTable.WithHasHeader(true).WithData(func() [][]string {
		rows := [][]string{{"Language", "Count", "Percentage"}}

		// Sort the languages again for display purposes.
		sortedLanguages := make([]struct {
			Language string
			Count    int
		}, 0, len(languageData))

		for lang, count := range languageData {
			sortedLanguages = append(sortedLanguages, struct {
				Language string
				Count    int
			}{lang, count})
		}

		sort.Slice(sortedLanguages, func(i, j int) bool {
			return sortedLanguages[i].Count > sortedLanguages[j].Count
		})

		// Calculate and add the percentage for each language.
		for _, langData := range sortedLanguages {
			percentage := float64(langData.Count) / float64(totalRepos) * 100
			rows = append(rows, []string{langData.Language, fmt.Sprintf("%d", langData.Count), fmt.Sprintf("%.2f%%", percentage)})
		}

		return rows
	}()).Render()

	return nil
}
