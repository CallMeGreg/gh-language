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
	org, _ := cmd.Flags().GetString("org")
	enterprise, _ := cmd.Flags().GetString("enterprise")
	repoLimit, _ := cmd.Flags().GetInt("repo-limit")
	orgLimit, _ := cmd.Flags().GetInt("org-limit")
	top := top_flag // Reuse the root command flag for top

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
		spinnerInfo, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Processing organization: %s", org))
		repos, err := FetchRepositories(org, repoLimit)
		if err != nil {
			spinnerInfo.Fail("Failed to process organization")
			return err
		}
		spinnerInfo.Success(fmt.Sprintf("Successfully processed organization: %s", org))

		totalRepos += len(repos)

		for _, repo := range repos {
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

			for lang := range repoLanguages {
				languageMapPerYear[time.Now().Year()][lang]++
			}
		}
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

func init() {
	RootCmd.AddCommand(trendCmd)
	trendCmd.Flags().String("org", "", "Organization name")
	trendCmd.Flags().String("enterprise", "", "Enterprise name")
	trendCmd.Flags().Int("repo-limit", 10, "The maximum number of repositories to evaluate per organization")
	trendCmd.Flags().Int("org-limit", 5, "The maximum number of organizations to evaluate for an enterprise")
	trendCmd.MarkFlagsMutuallyExclusive("org", "enterprise")
}
