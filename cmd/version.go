package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// Current version of the application
const Version = "v2.0.1"

// GitHub repository information
const (
	RepoOwner = "CallMeGreg"
	RepoName  = "gh-language"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.Info.Printf("gh-language version %s\n", Version)
	},
}

// CheckForUpdates checks if there's a newer version available
func CheckForUpdates() {
	client := &http.Client{Timeout: 5 * time.Second}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	resp, err := client.Get(url)
	if err != nil {
		// Silently fail - don't block the user if we can't check for updates
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Silently fail - don't block the user if API is unavailable
		return
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		// Silently fail - don't block the user if we can't parse the response
		return
	}

	if release.TagName != "" && release.TagName != Version {
		// Compare versions - if latest tag is different from current version, show warning
		if isNewerVersion(release.TagName, Version) {
			pterm.Warning.Printf("A newer version (%s) is available. Upgrade with: gh extension upgrade %s/%s\n",
				release.TagName, RepoOwner, RepoName)
			pterm.Println() // Add spacing
		}
	}
}

// isNewerVersion compares two version strings and returns true if latest is newer than current
func isNewerVersion(latest, current string) bool {
	// Simple version comparison - strip 'v' prefix and compare
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// For simplicity, just do string comparison - this works for most semantic versions
	// In a production environment, you might want to use a proper semver library
	return latest != current && latest > current
}

func init() {
	// Add version command as a hidden command to the root
	RootCmd.AddCommand(versionCmd)
}
