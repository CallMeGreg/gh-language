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
	Short: "Analyze language data by bytes",
	RunE:  runData,
}

func runData(cmd *cobra.Command, args []string) error {
	org, _ := cmd.Flags().GetString("org")
	unit, _ := cmd.Flags().GetString("unit")
	limit := limit_flag       // Reuse the root command flag for limit
	top := top_flag           // Reuse the root command flag for top
	language := language_flag // Reuse the root command flag for language

	if org == "" {
		return fmt.Errorf("--org flag is required")
	}

	if unit != "bytes" && unit != "kilobytes" && unit != "megabytes" && unit != "gigabytes" {
		return fmt.Errorf("invalid unit specified. Options are: bytes, kilobytes, megabytes, gigabytes")
	}

	pterm.DefaultSection.Println(fmt.Sprintf("Fetching repositories for organization: %s", org))

	repos, err := FetchRepositories(org, limit)
	if err != nil {
		return err
	}

	languageData := make(map[string]int)

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

		for lang, bytes := range repoLanguages {
			languageData[lang] += bytes
		}
	}

	progressBar.Stop()

	// Filter by specific language if --language flag is set
	if language != "" {
		filteredLanguageData := make(map[string]int)
		for lang, bytes := range languageData {
			if lang == language {
				filteredLanguageData[lang] = bytes
			}
		}
		languageData = filteredLanguageData
	}

	// Respect the --top flag
	if top > 0 {
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

		topLanguages := make(map[string]int)
		for i := 0; i < top && i < len(sortedLanguages); i++ {
			topLanguages[sortedLanguages[i].Language] = sortedLanguages[i].Bytes
		}

		languageData = topLanguages
	}

	pterm.DefaultTable.WithHasHeader(true).WithData(func() [][]string {
		rows := [][]string{{"Language", unit, "Percentage"}}

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

		for _, langData := range sortedLanguages {
			percentage := int(float64(languageData[langData.Language]) / float64(totalBytes) * 100)
			rows = append(rows, []string{langData.Language, fmt.Sprintf("%.2f", langData.Value), fmt.Sprintf("%d%%", percentage)})
		}

		return rows
	}()).Render()

	return nil
}

func init() {
	RootCmd.AddCommand(dataCmd)
	dataCmd.Flags().String("org", "", "Organization name")
	dataCmd.Flags().String("unit", "megabytes", "Unit to display language data (options: bytes, kilobytes, megabytes, gigabytes)")
}
