package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
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

// trendIndicator returns a colored arrow symbol and signed change string
// representing the year-over-year direction for a language.
func trendIndicator(current, previous int) (string, string) {
	diff := current - previous
	switch {
	case diff > 0:
		return pterm.Green("▲"), pterm.Green(fmt.Sprintf("+%d", diff))
	case diff < 0:
		return pterm.Red("▼"), pterm.Red(fmt.Sprintf("%d", diff))
	default:
		return pterm.Gray("●"), pterm.Gray("0")
	}
}

// topLanguageNames returns the names of the top N languages sorted by total count descending.
func topLanguageNames(trendData map[string]int, language string, top int) []string {
	type langCount struct {
		Language string
		Count    int
	}
	sorted := make([]langCount, 0, len(trendData))
	for lang, count := range trendData {
		if language != "" && lang != language {
			continue
		}
		sorted = append(sorted, langCount{lang, count})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Count > sorted[j].Count })

	limit := len(sorted)
	if top > 0 && top < limit {
		limit = top
	}
	names := make([]string, limit)
	for i := 0; i < limit; i++ {
		names[i] = sorted[i].Language
	}
	return names
}

// graphColors returns a cycling list of asciigraph colors for series lines.
func graphColors() []asciigraph.AnsiColor {
	return []asciigraph.AnsiColor{
		asciigraph.Red,
		asciigraph.Green,
		asciigraph.Yellow,
		asciigraph.Blue,
		asciigraph.Cyan,
		asciigraph.White,
	}
}

func runTrend(cmd *cobra.Command, args []string) error {
	org := org_flag
	enterprise := enterprise_flag
	repoLimit := repo_limit_flag
	orgLimit := org_limit_flag
	top := top_flag
	language := language_flag
	hostname := github_enterprise_server_url_flag

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
		orgs, err = FetchOrganizations(enterprise, orgLimit, hostname)
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
		totalReposInOrg, err := CountRepositoriesGraphQL(org, hostname)
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
		repos, err := FetchRepositoriesGraphQL(org, repoLimit, totalReposInOrg, hostname)
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
	pterm.Println()
	pterm.Info.Println(fmt.Sprintf("Total number of repositories analyzed: %d", totalRepos))
	pterm.Println()

	// Extract years and sort in ascending order (oldest first).
	years := make([]int, 0, len(languageMapPerYear))
	for year := range languageMapPerYear {
		years = append(years, year)
	}
	sort.Ints(years)

	if codeql_flag {
		for year, langMap := range languageMapPerYear {
			languageMapPerYear[year] = IsCodeQLLanguage(langMap)
		}
	}

	// Determine the top languages to focus on.
	topLangs := topLanguageNames(trendData, language, top)

	// ── Section 1: Trend Summary Table ──────────────────────────────
	// Shows each language with its latest-year count, year-over-year change,
	// and a trend indicator arrow.
	renderTrendSummary(languageMapPerYear, years, topLangs, totalRepos)

	// ── Section 2: Horizontal Bar Chart ─────────────────────────────
	// Visual bar chart for the most recent year's top languages.
	if len(years) > 0 {
		renderBarChart(languageMapPerYear, years, topLangs)
	}

	// ── Section 3: Multi-series Line Graph ──────────────────────────
	// ASCII line chart showing how each top language's count changes over time.
	if len(years) >= 2 {
		renderLineGraph(languageMapPerYear, years, topLangs)
	}

	// ── Section 4: Year-by-Year Detail Tables ───────────────────────
	// Detailed per-year tables with trend indicators compared to the prior year.
	renderYearTables(languageMapPerYear, years, topLangs, totalRepos, language, top)

	return nil
}

// renderTrendSummary displays a summary table with trend direction for each language.
func renderTrendSummary(languageMapPerYear map[int]map[string]int, years []int, topLangs []string, totalRepos int) {
	pterm.DefaultSection.Println("Language Trend Summary")

	if len(years) == 0 {
		pterm.Warning.Println("No data available.")
		return
	}

	latestYear := years[len(years)-1]

	rows := [][]string{{"Language", "Latest Count", "Trend", "YoY Change", "Percentage"}}
	for _, lang := range topLangs {
		latestCount := languageMapPerYear[latestYear][lang]

		// Find the previous year's count.
		var prevCount int
		if len(years) >= 2 {
			prevCount = languageMapPerYear[years[len(years)-2]][lang]
		}

		arrow, change := trendIndicator(latestCount, prevCount)
		percentage := 0
		if totalRepos > 0 {
			percentage = int(float64(latestCount) / float64(totalRepos) * 100)
		}

		rows = append(rows, []string{
			lang,
			fmt.Sprintf("%d", latestCount),
			arrow,
			change,
			fmt.Sprintf("%d%%", percentage),
		})
	}
	pterm.DefaultTable.WithHasHeader(true).WithData(rows).Render()
}

// renderBarChart displays a horizontal bar chart for the most recent year.
func renderBarChart(languageMapPerYear map[int]map[string]int, years []int, topLangs []string) {
	latestYear := years[len(years)-1]
	pterm.DefaultSection.Println(fmt.Sprintf("Top Languages — %d (Bar Chart)", latestYear))

	// Build bars from the top languages in the most recent year.
	barColors := []*pterm.Style{
		pterm.NewStyle(pterm.FgCyan),
		pterm.NewStyle(pterm.FgGreen),
		pterm.NewStyle(pterm.FgYellow),
		pterm.NewStyle(pterm.FgMagenta),
		pterm.NewStyle(pterm.FgRed),
		pterm.NewStyle(pterm.FgBlue),
		pterm.NewStyle(pterm.FgWhite),
	}

	bars := pterm.Bars{}
	for i, lang := range topLangs {
		count := languageMapPerYear[latestYear][lang]
		if count == 0 {
			continue
		}
		bars = append(bars, pterm.Bar{
			Label: lang,
			Value: count,
			Style: barColors[i%len(barColors)],
		})
	}

	if len(bars) > 0 {
		pterm.DefaultBarChart.
			WithHorizontal(true).
			WithShowValue(true).
			WithBars(bars).
			Render()
	}
}

// renderLineGraph displays a multi-series ASCII line graph showing language trends over time.
func renderLineGraph(languageMapPerYear map[int]map[string]int, years []int, topLangs []string) {
	pterm.DefaultSection.Println("Language Trends Over Time (Line Graph)")

	// Limit to a manageable number of series for readability.
	maxSeries := 6
	if len(topLangs) < maxSeries {
		maxSeries = len(topLangs)
	}
	langs := topLangs[:maxSeries]

	// Build data series: each series is a slice of float64 counts per year (ascending).
	allSeries := make([][]float64, len(langs))
	for i, lang := range langs {
		series := make([]float64, len(years))
		for j, year := range years {
			series[j] = float64(languageMapPerYear[year][lang])
		}
		allSeries[i] = series
	}

	// Build x-axis label caption showing year markers.
	yearLabels := make([]string, len(years))
	for i, y := range years {
		yearLabels[i] = fmt.Sprintf("%d", y)
	}

	colors := graphColors()
	seriesColors := make([]asciigraph.AnsiColor, len(langs))
	for i := range langs {
		seriesColors[i] = colors[i%len(colors)]
	}

	graph := asciigraph.PlotMany(allSeries,
		asciigraph.Height(15),
		asciigraph.Caption(strings.Join(yearLabels, "  →  ")),
		asciigraph.SeriesColors(seriesColors...),
		asciigraph.SeriesLegends(langs...),
	)
	pterm.Println(graph)
	pterm.Println()
}

// renderYearTables displays detailed per-year tables with trend indicators.
func renderYearTables(languageMapPerYear map[int]map[string]int, years []int, topLangs []string, totalRepos int, language string, top int) {
	pterm.DefaultSection.Println("Year-by-Year Breakdown")

	// Iterate years in descending order for the detail tables.
	for idx := len(years) - 1; idx >= 0; idx-- {
		year := years[idx]
		pterm.DefaultSection.WithLevel(2).Println(fmt.Sprintf("Year: %d", year))

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

		rows := [][]string{{"Language", "Count", "Percentage", "Trend", "YoY Change"}}
		for i, langData := range sortedLanguages {
			if top > 0 && i >= top {
				break
			}
			percentage := 0
			if totalRepos > 0 {
				percentage = int(float64(langData.Count) / float64(totalRepos) * 100)
			}

			arrow := ""
			change := ""
			if idx > 0 {
				prevYear := years[idx-1]
				prevCount := languageMapPerYear[prevYear][langData.Language]
				arrow, change = trendIndicator(langData.Count, prevCount)
			}

			rows = append(rows, []string{
				langData.Language,
				fmt.Sprintf("%d", langData.Count),
				fmt.Sprintf("%d%%", percentage),
				arrow,
				change,
			})
		}

		pterm.DefaultTable.WithHasHeader(true).WithData(rows).Render()
	}
}
