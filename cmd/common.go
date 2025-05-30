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

func Yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[m"
}

// FetchRepositories fetches repositories for a given organization and limit.
func FetchRepositories(org string, limit int) ([]struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}, error) {
	if org == "" {
		return nil, fmt.Errorf("no organization identified, please ensure you have access to the organization or enterprise provided")
	}

	client, err := api.DefaultRESTClient()
	if err != nil {
		pterm.Error.Println("Failed to create REST client:", err)
		return nil, err
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

// FetchOrganizations fetches organizations for a given enterprise using the GitHub GraphQL API.
func FetchOrganizations(enterprise string, orgLimit int) ([]string, error) {
	if enterprise == "" {
		return nil, fmt.Errorf("--enterprise flag is required")
	}

	query := `{
		enterprise(slug: "` + enterprise + `") {
			organizations(first: ` + fmt.Sprintf("%d", orgLimit) + `) {
				nodes {
					login
				}
			}
		}
	}`

	response, _, err := gh.Exec("api", "graphql", "-f", "query="+query)
	if err != nil {
		pterm.Error.Println("Failed to fetch organizations for enterprise:", err)
		return nil, err
	}

	var result struct {
		Data struct {
			Enterprise struct {
				Organizations struct {
					Nodes []struct {
						Login string `json:"login"`
					}
				} `json:"organizations"`
			} `json:"enterprise"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response.Bytes(), &result); err != nil {
		pterm.Error.Println("Failed to parse organizations data:", err)
		return nil, err
	}

	var orgs []string
	for _, org := range result.Data.Enterprise.Organizations.Nodes {
		orgs = append(orgs, org.Login)
	}

	return orgs, nil
}

// IsCodeQLLanguage filters a map of languages to only include CodeQL-supported languages.
func IsCodeQLLanguage(languageData map[string]int) map[string]int {
	allowedLanguages := map[string]bool{
		"C": true, "C++": true, "C#": true, "Go": true, "HTML": true, "Java": true, "Kotlin": true,
		"JavaScript": true, "Python": true, "Ruby": true, "Swift": true, "TypeScript": true,
		"Vue": true,
	}
	for lang := range languageData {
		if !allowedLanguages[lang] {
			delete(languageData, lang)
		}
	}
	return languageData
}

// PrintInfo prints an informational message with pterm.
func PrintInfo(message string) {
	pterm.Info.Println(message)
}

// PrintInfoWithFormat formats a message and prints it as informational.
func PrintInfoWithFormat(format string, args ...interface{}) {
	pterm.Info.Println(fmt.Sprintf(format, args...))
}

// PrintIndexingEnterprise prints a message about indexing organizations for an enterprise.
func PrintIndexingEnterprise(enterprise string) {
	PrintInfoWithFormat("Indexing organizations for enterprise: %s", enterprise)
}

// PrintTotalRepositories prints the total number of repositories analyzed.
func PrintTotalRepositories(total int) {
	pterm.Println() // Add a new line
	PrintInfoWithFormat("Total number of repositories analyzed: %d", total)
	pterm.Println() // Add a new line
}

// ValidateFlags checks if the required flags are set and returns an error if not.
func ValidateFlags(org, enterprise string) error {
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
