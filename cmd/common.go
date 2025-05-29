package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh"
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
		return nil, fmt.Errorf("--org flag is required")
	}

	reposOutput, _, err := gh.Exec("api", fmt.Sprintf("orgs/%s/repos?per_page=%d", org, limit))
	if err != nil {
		pterm.Error.Println("Failed to fetch repositories:", err)
		return nil, err
	}

	var repos []struct {
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(reposOutput.Bytes(), &repos); err != nil {
		pterm.Error.Println("Failed to parse repositories data:", err)
		return nil, err
	}

	return repos, nil
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
