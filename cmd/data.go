package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cli/go-gh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type LanguageData struct {
	Language string `json:"language"`
	Bytes    int    `json:"bytes"`
}

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Analyze the programming languages used in repos across an enterprise or organization based on bytes of data",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runData(cmd, args)
	},
}

func runData(cmd *cobra.Command, args []string) error {
	org := org_flag
	enterprise := enterprise_flag
	repoLimit := repo_limit_flag
	orgLimit := org_limit_flag
	top := top_flag
	language := language_flag
	unit, _ := cmd.Flags().GetString("unit")

	if err := ValidateFlags(org, enterprise); err != nil {
		return err
	}

	if unit != "bytes" && unit != "kilobytes" && unit != "megabytes" && unit != "gigabytes" {
		// Validate the unit flag to ensure it is one of the allowed values.
		return fmt.Errorf("invalid unit specified. Options are: bytes, kilobytes, megabytes, gigabytes")
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

		// Analyze each repository for language usage.
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

			// Update the language data map with the parsed data.
			for lang, bytes := range repoLanguages {
				if bytes > 0 {
					languageData[lang] += bytes
				}
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
		for lang, bytes := range languageData {
			if lang == language {
				filteredLanguageData[lang] = bytes
			}
		}
		languageData = filteredLanguageData
	}

	// Respect the --top flag by limiting the number of languages displayed.
	if top > 0 {
		// Sort the languages by their usage in descending order.
		sortedLanguages := make([]struct {
			Language string
			Bytes    int
		}, 0, len(languageData))

		for lang, bytes := range languageData {
			sortedLanguages = append(sortedLanguages, struct {
				Language string
				Bytes    int
			}{lang, bytes})
		}

		sort.Slice(sortedLanguages, func(i, j int) bool {
			return sortedLanguages[i].Bytes > sortedLanguages[j].Bytes
		})

		// Select the top N languages based on the --top flag.
		topLanguages := make(map[string]int)
		for i := 0; i < top && i < len(sortedLanguages); i++ {
			topLanguages[sortedLanguages[i].Language] = sortedLanguages[i].Bytes
		}

		languageData = topLanguages
	}

	// If the CodeQL flag is set, filter the language data to include only CodeQL-supported languages.
	if codeql_flag {
		languageData = IsCodeQLLanguage(languageData)
	}

	// Render the language data as a table with percentages.
	pterm.DefaultTable.WithHasHeader(true).WithData(func() [][]string {
		rows := [][]string{{"Language", unit, "Percentage"}}

		// Sort the languages again for display purposes.
		sortedLanguages := make([]struct {
			Language string
			Value    float64
		}, 0, len(languageData))

		var totalBytes int
		for _, bytes := range languageData {
			totalBytes += bytes
		}

		for lang, bytes := range languageData {
			var value float64
			switch unit {
			case "bytes":
				value = float64(bytes)
			case "kilobytes":
				value = float64(bytes) / 1024
			case "megabytes":
				value = float64(bytes) / 1024 / 1024
			case "gigabytes":
				value = float64(bytes) / 1024 / 1024 / 1024
			}
			sortedLanguages = append(sortedLanguages, struct {
				Language string
				Value    float64
			}{lang, value})
		}

		sort.Slice(sortedLanguages, func(i, j int) bool {
			return sortedLanguages[i].Value > sortedLanguages[j].Value
		})

		// Add each language and its usage to the table rows.
		for _, langData := range sortedLanguages {
			percentage := int(float64(languageData[langData.Language]) / float64(totalBytes) * 100)
			rows = append(rows, []string{langData.Language, fmt.Sprintf("%d", int(langData.Value)), fmt.Sprintf("%d%%", percentage)})
		}

		return rows
	}()).Render()

	return nil
}

func init() {
	dataCmd.Flags().String("unit", "bytes", "Specify the unit for language data (bytes, kilobytes, megabytes, gigabytes)")
}
