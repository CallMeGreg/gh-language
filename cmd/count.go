package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cli/go-gh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Analyze the count of programming languages used in repos across an organization",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runCount(cmd, args)
	},
}

func runCount(cmd *cobra.Command, args []string) error {
	org, _ := cmd.Flags().GetString("org")
	enterprise, _ := cmd.Flags().GetString("enterprise")
	repoLimit, _ := cmd.Flags().GetInt("repo-limit")
	orgLimit, _ := cmd.Flags().GetInt("org-limit")
	top := top_flag           // Reuse the root command flag for top
	language := language_flag // Reuse the root command flag for language

	if org == "" && enterprise == "" {
		return fmt.Errorf("either --org or --enterprise flag is required")
	}

	var orgs []string

	if enterprise != "" {
		pterm.Info.Println(fmt.Sprintf("Organization limit: %d, Repository limit: %d, Top languages limit: %d", orgLimit, repoLimit, top))
		pterm.Info.Println(fmt.Sprintf("Indexing organizations for enterprise: %s", enterprise))
		var err error
		orgs, err = FetchOrganizations(enterprise, orgLimit)
		if err != nil {
			return err
		}
	} else {
		pterm.Info.Println(fmt.Sprintf("Repository limit: %d, Top languages limit: %d", repoLimit, top))
		orgs = []string{org}
	}

	languageData := make(map[string]int)

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

			for lang := range repoLanguages {
				languageData[lang]++
			}
		}

		progressBar.Stop()
	}

	pterm.Println() // Add a new line
	pterm.Info.Println(fmt.Sprintf("Total number of repositories analyzed: %d", totalRepos))

	pterm.Println() // Add a new line

	// Filter by specific language if --language flag is set
	if language != "" {
		filteredLanguageData := make(map[string]int)
		for lang, count := range languageData {
			if lang == language {
				filteredLanguageData[lang] = count
			}
		}
		languageData = filteredLanguageData
	}

	// Respect the --top flag
	if top > 0 {
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

		topLanguages := make(map[string]int)
		for i := 0; i < top && i < len(sortedLanguages); i++ {
			topLanguages[sortedLanguages[i].Language] = sortedLanguages[i].Count
		}

		languageData = topLanguages
	}

	// Update percentage calculation
	pterm.DefaultTable.WithHasHeader(true).WithData(func() [][]string {
		rows := [][]string{{"Language", "Count", "Percentage"}}

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

		for _, langData := range sortedLanguages {
			percentage := int(float64(langData.Count) / float64(totalRepos) * 100)
			rows = append(rows, []string{langData.Language, fmt.Sprintf("%d", langData.Count), fmt.Sprintf("%d%%", percentage)})
		}

		return rows
	}()).Render()

	return nil
}

func init() {
	RootCmd.AddCommand(countCmd)
	countCmd.Flags().String("org", "", "Organization name")
	countCmd.Flags().String("enterprise", "", "Enterprise name")
	countCmd.Flags().Int("repo-limit", 10, "The maximum number of repositories to evaluate per organization")
	countCmd.Flags().Int("org-limit", 5, "The maximum number of organizations to evaluate for an enterprise")
	countCmd.MarkFlagsMutuallyExclusive("org", "enterprise")
}
