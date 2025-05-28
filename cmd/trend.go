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
	limit := limit_flag // Reuse the root command flag for limit
	top := top_flag     // Reuse the root command flag for top

	repos, err := FetchRepositories(org, limit)
	if err != nil {
		return err
	}

	if len(repos) > limit {
		ShowProgressBar(limit, "Fetching repositories")
	}

	languageMapPerYear := make(map[int]map[string]int)

	progressBar, _ := pterm.DefaultProgressbar.WithTotal(len(repos)).WithTitle("Processing repositories").Start()

	for _, repo := range repos {
		progressBar.UpdateTitle("Fetching language data for repositories...")
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

		t, err := time.Parse(GITHUB_TIMESTAMP_LAYOUT, repo.CreatedAt)
		if err != nil {
			pterm.Warning.Println(fmt.Sprintf("Skipping repository %s due to date parsing error: %s", repo.Name, err))
			continue
		}
		year := t.Year()
		if _, ok := languageMapPerYear[year]; !ok {
			languageMapPerYear[year] = make(map[string]int)
		}
		for lang := range repoLanguages {
			languageMapPerYear[year][lang]++
		}
	}

	progressBar.Stop()

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
			percentage := int(float64(langData.Count) / float64(len(repos)) * 100)
			rows = append(rows, []string{langData.Language, fmt.Sprintf("%d", langData.Count), fmt.Sprintf("%d%%", percentage)})
		}

		pterm.DefaultTable.WithHasHeader(true).WithData(rows).Render()
	}

	return nil
}

func init() {
	RootCmd.AddCommand(trendCmd)
	trendCmd.Flags().String("org", "", "Organization name")
}
