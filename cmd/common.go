package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/pterm/pterm"
)

func Red(s string) string {
	return "\x1b[31m" + s + "\x1b[m"
}

// FetchOrganizations fetches organizations for a given enterprise using the GitHub GraphQL API.
func FetchOrganizations(enterprise string, orgLimit int) ([]string, error) {
	if enterprise == "" {
		return nil, fmt.Errorf("--enterprise flag is required")
	}

	const maxPerPage = 100
	var orgs []string
	var cursor *string
	fetched := 0

	for {
		remaining := orgLimit - fetched
		if remaining > maxPerPage {
			remaining = maxPerPage
		}

		query := `{
			enterprise(slug: "` + enterprise + `") {
				organizations(first: ` + fmt.Sprintf("%d", remaining) + `, after: ` + formatCursor(cursor) + `) {
					nodes {
						login
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}`

		response, stderr, err := gh.Exec("api", "graphql", "-f", "query="+query)
		if err != nil {
			pterm.Error.Printf("Failed to fetch organizations for enterprise '%s': %v\n", enterprise, err)
			pterm.Error.Printf("GraphQL query: %s\n", query)
			pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
			return nil, err
		}

		var result struct {
			Data struct {
				Enterprise struct {
					Organizations struct {
						Nodes []struct {
							Login string `json:"login"`
						}
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"organizations"`
				} `json:"enterprise"`
			} `json:"data"`
		}

		if err := json.Unmarshal(response.Bytes(), &result); err != nil {
			pterm.Error.Printf("Failed to parse organizations data for enterprise '%s': %v\n", enterprise, err)
			return nil, err
		}

		for _, org := range result.Data.Enterprise.Organizations.Nodes {
			orgs = append(orgs, org.Login)
			fetched++
			if fetched >= orgLimit {
				return orgs, nil
			}
		}

		if !result.Data.Enterprise.Organizations.PageInfo.HasNextPage {
			break
		}
		cursor = &result.Data.Enterprise.Organizations.PageInfo.EndCursor
	}

	return orgs, nil
}

func formatCursor(cursor *string) string {
	if cursor == nil {
		return "null"
	}
	return `"` + *cursor + `"`
}

// FetchRepositories fetches repositories for a given organization and limit.
func FetchRepositories(client *api.RESTClient, org string, limit int) ([]struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}, error) {
	if org == "" {
		return nil, fmt.Errorf("no organization identified, please ensure you have access to the organization or enterprise provided")
	}

	var allRepos []struct {
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}

	requestPath := fmt.Sprintf("orgs/%s/repos?per_page=100", org)
	fetched := 0

	for {
		response, err := client.Request(http.MethodGet, requestPath, nil)
		if err != nil {
			pterm.Error.Println("Failed to fetch repositories:", err)
			return nil, err
		}

		// Check rate limit headers
		remaining := response.Header.Get("X-RateLimit-Remaining")
		reset := response.Header.Get("X-RateLimit-Reset")
		if remaining == "0" {
			resetTime, _ := strconv.Atoi(reset)
			waitDuration := time.Until(time.Unix(int64(resetTime), 0))
			pterm.Warning.Printf("Rate limit exceeded. Waiting for %v...\n", waitDuration)
			time.Sleep(waitDuration)
			continue
		}

		var repos []struct {
			Name      string `json:"name"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.NewDecoder(response.Body).Decode(&repos); err != nil {
			pterm.Error.Println("Failed to parse repositories data:", err)
			return nil, err
		}
		response.Body.Close()

		allRepos = append(allRepos, repos...)
		fetched += len(repos)
		if fetched >= limit || len(repos) == 0 {
			break
		}

		// Find next page URL from Link header
		linkHeader := response.Header.Get("Link")
		nextPageURL := findNextPage(linkHeader)
		if nextPageURL == "" {
			break
		}
		requestPath = nextPageURL
	}

	if len(allRepos) > limit {
		allRepos = allRepos[:limit]
	}

	return allRepos, nil
}

// findNextPage extracts the next page URL from the Link header.
func findNextPage(linkHeader string) string {
	var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)
	matches := linkRE.FindStringSubmatch(linkHeader)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ShowProgressBar displays a progress bar.
func ShowProgressBar(total int, title string) {
	progressBar, _ := pterm.DefaultProgressbar.WithTotal(total).WithTitle(title).Start()
	progressBar.Increment()
	progressBar.Stop()
}

// IsCodeQLLanguage filters a map of languages to only include CodeQL-supported languages.
func IsCodeQLLanguage(languageData map[string]int) map[string]int {
	allowedLanguages := map[string]bool{
		"C": true, "C++": true, "C#": true, "Go": true, "HTML": true, "Java": true, "Kotlin": true,
		"JavaScript": true, "Python": true, "Ruby": true, "Swift": true, "TypeScript": true,
		"Vue": true,
	}
	filteredLanguages := make(map[string]int)
	for lang, count := range languageData {
		if allowedLanguages[lang] {
			filteredLanguages[lang] = count
		}
	}
	return filteredLanguages
}

// PrintInfo prints an informational message with pterm.
func PrintInfo(message string) {
	pterm.Info.Println(message)
}

// PrintInfoWithFormat formats a message and prints it as informational.
func PrintInfoWithFormat(format string, args ...interface{}) {
	pterm.Info.Println(fmt.Sprintf(format, args...))
}

// StartIndexingEnterpriseSpinner starts a spinner for indexing organizations for an enterprise.
func StartIndexingEnterpriseSpinner(enterprise string) (*pterm.SpinnerPrinter, error) {
	return pterm.DefaultSpinner.Start(fmt.Sprintf("Indexing organizations for enterprise: %s", enterprise))
}

// PrintTotalOrganizations prints the total number of organizations found.
func PrintTotalOrganizations(total int) {
	PrintInfoWithFormat("Total number of organizations found: %d", total)
}

// PrintTotalRepositories prints the total number of repositories analyzed.
func PrintTotalRepositories(total int) {
	pterm.Println() // Add a new line
	PrintInfoWithFormat("Total number of repositories analyzed: %d", total)
	pterm.Println() // Add a new line
}

// ValidateFlags checks if the required flags are set and returns an error if not.
func ValidateFlags(org, enterprise string) error {
	// Check for updates before validation (only once per command execution)
	CheckForUpdates()

	if org == "" && enterprise == "" {
		return fmt.Errorf("either --org or --enterprise flag is required")
	}
	return nil
}

// GetLanguageFilter determines the language filter info based on flags.
func GetLanguageFilter(codeqlFlag bool, language string, top int) string {
	if codeqlFlag {
		return "CodeQL language filter applied"
	} else if language != "" {
		return fmt.Sprintf("Language filter: %s", language)
	}
	return fmt.Sprintf("Top languages limit: %d", top)
}

// FetchLanguages fetches the programming languages used in a repository.
func FetchLanguages(client *api.RESTClient, org, repo string) (map[string]int, error) {
	if org == "" || repo == "" {
		return nil, fmt.Errorf("organization and repository names are required")
	}

	requestPath := fmt.Sprintf("repos/%s/%s/languages", org, repo)
	var languages map[string]int

	for {
		response, err := client.Request(http.MethodGet, requestPath, nil)
		if err != nil {
			pterm.Error.Println("Failed to fetch languages for repository:", err)
			return nil, err
		}

		// Check rate limit headers
		remaining := response.Header.Get("X-RateLimit-Remaining")
		reset := response.Header.Get("X-RateLimit-Reset")
		if remaining == "0" {
			resetTime, _ := strconv.Atoi(reset)
			waitDuration := time.Until(time.Unix(int64(resetTime), 0))
			pterm.Warning.Printf("Rate limit exceeded. Waiting for %v...\n", waitDuration)
			time.Sleep(waitDuration)
			continue
		}

		if err := json.NewDecoder(response.Body).Decode(&languages); err != nil {
			pterm.Error.Println("Failed to parse language data:", err)
			return nil, err
		}
		response.Body.Close()
		break // No pagination for language data, so exit loop
	}

	return languages, nil
}

// FetchRepositoriesGraphQL fetches repositories with languages for a given organization using GraphQL API with pagination.
func FetchRepositoriesGraphQL(org string, limit int, totalRepos int) ([]struct {
	Name      string              `json:"name"`
	CreatedAt string              `json:"created_at"`
	Languages map[string]struct{} `json:"languages"`
}, error) {
	if org == "" {
		return nil, fmt.Errorf("no organization identified, please ensure you have access to the organization or enterprise provided")
	}

	const maxPerPage = 100
	var allRepos []struct {
		Name      string              `json:"name"`
		CreatedAt string              `json:"created_at"`
		Languages map[string]struct{} `json:"languages"`
	}

	var cursor *string
	fetched := 0
	pageCount := 0

	// Use progress bar since we know the total number of repos available
	progressTarget := limit
	if totalRepos < limit {
		progressTarget = totalRepos
	}
	progressBar, _ := pterm.DefaultProgressbar.WithTotal(progressTarget).WithTitle("Fetching repositories and their languages").Start()

	for {
		remaining := limit - fetched
		if remaining <= 0 {
			break
		}
		if remaining > maxPerPage {
			remaining = maxPerPage
		}

		pageCount++

		query := fmt.Sprintf(`{
			organization(login: "%s") {
				repositories(first: %d, after: %s) {
					nodes {
						name
						createdAt
						languages(first: 100) {
							nodes {
								name
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}`, org, remaining, formatCursor(cursor))

		response, stderr, err := gh.Exec("api", "graphql", "-f", "query="+query)
		if err != nil {
			progressBar.Stop()
			pterm.Error.Printf("Failed to fetch repositories for organization '%s': %v\n", org, err)
			pterm.Error.Printf("GraphQL query: %s\n", query)
			pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
			return nil, err
		}

		var result struct {
			Data struct {
				Organization struct {
					Repositories struct {
						Nodes []struct {
							Name      string `json:"name"`
							CreatedAt string `json:"createdAt"`
							Languages struct {
								Nodes []struct {
									Name string `json:"name"`
								} `json:"nodes"`
							} `json:"languages"`
						} `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"repositories"`
				} `json:"organization"`
			} `json:"data"`
		}

		if err := json.Unmarshal(response.Bytes(), &result); err != nil {
			progressBar.Stop()
			pterm.Error.Printf("Failed to parse repositories data for organization '%s': %v\n", org, err)
			return nil, err
		}

		// Check if organization exists
		if len(result.Data.Organization.Repositories.Nodes) == 0 && cursor == nil {
			progressBar.Stop()
			pterm.Warning.Printf("No repositories found for organization: %s\n", org)
			return allRepos, nil
		}

		// Process repositories from this page
		reposInThisPage := 0
		for _, repo := range result.Data.Organization.Repositories.Nodes {
			// Convert language nodes to map for compatibility
			languages := make(map[string]struct{})
			for _, lang := range repo.Languages.Nodes {
				languages[lang.Name] = struct{}{}
			}

			allRepos = append(allRepos, struct {
				Name      string              `json:"name"`
				CreatedAt string              `json:"created_at"`
				Languages map[string]struct{} `json:"languages"`
			}{
				Name:      repo.Name,
				CreatedAt: repo.CreatedAt,
				Languages: languages,
			})

			fetched++
			reposInThisPage++
			if fetched >= limit {
				break
			}
		}

		// Update progress bar with current progress after processing this page
		progressBar.Add(reposInThisPage)

		// Check if we've reached the limit or if there are no more pages
		if fetched >= limit || !result.Data.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		cursor = &result.Data.Organization.Repositories.PageInfo.EndCursor

	}

	progressBar.Stop()
	return allRepos, nil
}

// CountRepositoriesGraphQL counts the total number of repositories in an organization using GraphQL API.
func CountRepositoriesGraphQL(org string) (int, error) {
	if org == "" {
		return 0, fmt.Errorf("no organization identified, please ensure you have access to the organization or enterprise provided")
	}

	query := fmt.Sprintf(`{
		organization(login: "%s") {
			repositories {
				totalCount
			}
		}
	}`, org)

	response, stderr, err := gh.Exec("api", "graphql", "-f", "query="+query)
	if err != nil {
		pterm.Error.Printf("Failed to count repositories for organization '%s': %v\n", org, err)
		pterm.Error.Printf("GraphQL query: %s\n", query)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return 0, err
	}

	var result struct {
		Data struct {
			Organization struct {
				Repositories struct {
					TotalCount int `json:"totalCount"`
				} `json:"repositories"`
			} `json:"organization"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response.Bytes(), &result); err != nil {
		pterm.Error.Printf("Failed to parse repository count data for organization '%s': %v\n", org, err)
		return 0, err
	}

	return result.Data.Organization.Repositories.TotalCount, nil
}
