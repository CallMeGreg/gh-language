package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const GITHUB_TIMESTAMP_LAYOUT = "2006-01-02T15:04:05Z"

// MAX_GRAPH_SERIES limits the number of language series shown in the line graph for readability.
const MAX_GRAPH_SERIES = 10

var min_year_flag int
var max_year_flag int

func init() {
	trendCmd.Flags().IntVar(&min_year_flag, "min-year", 0, "Minimum year to include in the trend output")
	trendCmd.Flags().IntVar(&max_year_flag, "max-year", time.Now().Year()-1, "Maximum year to include in the trend output (defaults to last year)")
}

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
	languages := ParseLanguages(language)
	sorted := make([]langCount, 0, len(trendData))
	for lang, count := range trendData {
		if len(languages) > 0 && !MatchesLanguageFilter(lang, languages) {
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
		asciigraph.DarkOrange,
		asciigraph.BlueViolet,
		asciigraph.Coral,
		asciigraph.Chartreuse,
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

	if min_year_flag > 0 && max_year_flag > 0 && min_year_flag > max_year_flag {
		return fmt.Errorf("--min-year (%d) cannot be greater than --max-year (%d)", min_year_flag, max_year_flag)
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

	// Initialize a map to store number of repos per year.
	reposPerYear := make(map[int]int)

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
			reposPerYear[creationYear]++
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

	// Filter years based on min-year and max-year flags.
	if min_year_flag > 0 || max_year_flag > 0 {
		filtered := make([]int, 0, len(years))
		for _, y := range years {
			if min_year_flag > 0 && y < min_year_flag {
				continue
			}
			if max_year_flag > 0 && y > max_year_flag {
				continue
			}
			filtered = append(filtered, y)
		}
		years = filtered
	}

	if codeql_flag {
		for year, langMap := range languageMapPerYear {
			languageMapPerYear[year] = IsCodeQLLanguage(langMap)
		}
	}

	// Determine the top languages to focus on.
	topLangs := topLanguageNames(trendData, language, top)

	// ── Section 1: Multi-series Line Graph ──────────────────────────
	// ASCII line chart showing how each top language's count changes over time.
	if len(years) >= 2 {
		renderLineGraph(languageMapPerYear, years, topLangs)
	}

	// ── Section 2: Year-by-Year Detail Tables ───────────────────────
	// Detailed per-year tables with trend indicators compared to the prior year.
	renderYearTables(languageMapPerYear, years, topLangs, reposPerYear, language, top)

	return nil
}

// renderLineGraph displays a multi-series ASCII line graph showing language trends over time.
func renderLineGraph(languageMapPerYear map[int]map[string]int, years []int, topLangs []string) {
	pterm.DefaultSection.Println("Language Trends Over Time (Repo Count Created by Year)")

	// Rank languages by their count in the max (last) year, descending.
	maxYear := years[len(years)-1]
	maxYearData := languageMapPerYear[maxYear]

	type langCount struct {
		Language string
		Count    int
	}
	ranked := make([]langCount, 0, len(topLangs))
	for _, lang := range topLangs {
		ranked = append(ranked, langCount{lang, maxYearData[lang]})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].Count > ranked[j].Count })

	maxSeries := MAX_GRAPH_SERIES
	if len(ranked) < maxSeries {
		maxSeries = len(ranked)
	}
	langs := make([]string, maxSeries)
	for i := 0; i < maxSeries; i++ {
		langs[i] = ranked[i].Language
	}

	// Build data series: each series is a slice of float64 counts per year (ascending).
	allSeries := make([][]float64, len(langs))
	for i, lang := range langs {
		series := make([]float64, len(years))
		for j, year := range years {
			series[j] = float64(languageMapPerYear[year][lang])
		}
		allSeries[i] = series
	}

	// Build x-axis label caption showing first and last year.
	caption := fmt.Sprintf("%d → %d", years[0], years[len(years)-1])

	colors := graphColors()
	seriesColors := make([]asciigraph.AnsiColor, len(langs))
	for i := range langs {
		seriesColors[i] = colors[i%len(colors)]
	}

	graph := asciigraph.PlotMany(allSeries,
		asciigraph.Height(15),
		asciigraph.Width(len(years)*3),
		asciigraph.Precision(0),
		asciigraph.Caption(caption),
		asciigraph.SeriesColors(seriesColors...),
		asciigraph.SeriesLegends(langs...),
	)
	pterm.Println(graph)
	pterm.Println()
}

// renderYearTables displays detailed per-year tables with trend indicators.
func renderYearTables(languageMapPerYear map[int]map[string]int, years []int, topLangs []string, reposPerYear map[int]int, language string, top int) {
	pterm.DefaultSection.Println("Year-by-Year Breakdown")

	// Iterate years in descending order for the detail tables.
	for idx := len(years) - 1; idx >= 0; idx-- {
		year := years[idx]
		yearRepoCount := reposPerYear[year]
		pterm.DefaultSection.WithLevel(2).Println(fmt.Sprintf("Year: %d (%d repos)", year, yearRepoCount))

		sortedLanguages := make([]struct {
			Language string
			Count    int
		}, 0, len(languageMapPerYear[year]))

		languages := ParseLanguages(language)
		for lang, count := range languageMapPerYear[year] {
			if len(languages) > 0 && !MatchesLanguageFilter(lang, languages) {
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
			if yearRepoCount > 0 {
				percentage = int(float64(langData.Count) / float64(yearRepoCount) * 100)
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
