package cmd

import (
	"fmt"
	"sort"
	"time"

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
		spinnerEnterprise, _ := StartIndexingEnterpriseSpinner(enterprise)
		var err error
		orgs, err = FetchOrganizations(enterprise, orgLimit)
		if err != nil {
			spinnerEnterprise.Fail("Failed to index organizations for enterprise")
			return err
		}
		spinnerEnterprise.Success(fmt.Sprintf("Successfully indexed enterprise: %s", enterprise))
		PrintTotalOrganizations(len(orgs))
	} else {
		// Handle the case where only a single organization is provided.
		topLanguagesInfo := GetLanguageFilter(codeql_flag, language, top)
		PrintInfoWithFormat("Repository limit: %d, %s", repoLimit, topLanguagesInfo)
		orgs = []string{org}
	}

	// Initialize a map to store language data per year.
	languageMapPerYear := make(map[int]map[string]int)

	// Initialize trendData as a map to store language trends.
	trendData := make(map[string]int)

	var totalRepos int

	// Iterate over each organization to fetch repositories and analyze languages.
	for orgIndex, org := range orgs {
		// Start a spinner to indicate progress for indexing the organization.
		spinnerInfo, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Indexing organization: %s", org))

		// First, count the total number of repositories in the organization
		totalReposInOrg, err := CountRepositoriesGraphQL(org)
		if err != nil {
			// Stop the spinner and indicate failure if an error occurs.
			spinnerInfo.Fail("Failed to index organization")
			return err
		}

		if totalReposInOrg == 0 {
			// Stop the spinner and indicate a warning if no repositories are found.
			spinnerInfo.Warning(fmt.Sprintf("No repositories found for organization %d of %d: %s", orgIndex+1, len(orgs), org))
			continue
		}

		// Apply the repo limit to determine effective repository count
		effectiveRepoCount := totalReposInOrg
		if repoLimit < totalReposInOrg {
			effectiveRepoCount = repoLimit
		}

		// Stop the spinner and indicate success.
		spinnerInfo.Success(fmt.Sprintf("Successfully indexed organization %d of %d: %s (%d repositories, limited to %d)", orgIndex+1, len(orgs), org, totalReposInOrg, effectiveRepoCount))

		// Fetch repositories with languages using GraphQL API with progress bar.
		repos, err := FetchRepositoriesGraphQL(org, repoLimit, totalReposInOrg)
		if err != nil {
			return err
		}

		// Increment the total repository count.
		totalRepos += len(repos)

		// Analyze each repository for language usage and group by year.
		for _, repo := range repos {
			// Update the trend data map with the fetched data by incrementing the count.
			for lang := range repo.Languages {
				trendData[lang]++
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

			for lang := range repo.Languages {
				languageMapPerYear[creationYear][lang]++
			}
		}
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
