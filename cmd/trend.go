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
	Short: "Analyze the trend of programming languages used in repos across an organization over time",
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

	if org == "" && enterprise == "" {
		return fmt.Errorf("either --org or --enterprise flag is required")
	}

	var orgs []string
	pterm.Info.Println(fmt.Sprintf("Organization limit: %d, Repository limit: %d, Top languages limit: %d", orgLimit, repoLimit, top))

	if enterprise != "" {
		pterm.Info.Println(fmt.Sprintf("Indexing organizations for enterprise: %s", enterprise))
		var err error
		orgs, err = FetchOrganizations(enterprise, orgLimit)
		if err != nil {
			return err
		}
	} else {
		orgs = []string{org}
	}

	languageMapPerYear := make(map[int]map[string]int)

	var totalRepos int

	for _, org := range orgs {
		spinnerInfo, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Indexing organization: %s", org))

		repos, err := FetchRepositories(org, repoLimit)
		if err != nil {
			spinnerInfo.Fail("Failed to index organization")
			return err
		}

		if len(repos) == 0 {
			spinnerInfo.Warning(fmt.Sprintf("No repositories found for organization: %s", org))
			continue
		}

		spinnerInfo.Success(fmt.Sprintf("Successfully indexed organization: %s", org))
		progressBar, _ := pterm.DefaultProgressbar.WithTotal(len(repos)).WithTitle("Analyzing repositories").Start()

		totalRepos += len(repos)

		for _, repo := range repos {
			progressBar.Increment()

			output, _, err := gh.Exec("api", fmt.Sprintf("repos/%s/%s/languages", org, repo.Name))
			if err != nil {
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to error: %s", repo.Name, err))
				continue
			}

			var repoLanguages map[string]int
			if err := json.Unmarshal(output.Bytes(), &repoLanguages); err != nil {
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to parsing error: %s", repo.Name, err))
				continue
			}

			// Parse the repository's creation date
			createdAt, err := time.Parse(GITHUB_TIMESTAMP_LAYOUT, repo.CreatedAt)
			if err != nil {
				pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to invalid creation date: %s", repo.Name, err))
				continue
			}

			creationYear := createdAt.Year()
			if languageMapPerYear[creationYear] == nil {
				languageMapPerYear[creationYear] = make(map[string]int)
			}

			for lang := range repoLanguages {
				languageMapPerYear[creationYear][lang]++
			}
		}

		progressBar.Stop()
	}

	// Extract the keys (years) into a slice
	years := make([]int, 0, len(languageMapPerYear))
	for year := range languageMapPerYear {
		years = append(years, year)
	}

	// Sort the slice in ascending order
	sort.Ints(years)

	// Reverse the slice to get it in descending order
	for i, j := 0, len(years)-1; i < j; i, j = i+1, j-1 {
		years[i], years[j] = years[j], years[i]
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
