package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cli/go-gh"
)

type languageDetails struct {
	Size      int    `json:"size"`
	CreatedAt string `json:"createdAt"`
	Node      struct {
		Name string `json:"name"`
	} `json:"node"`
}

type Repo struct {
	Languages     []languageDetails `json:"languages"`
	CreatedAt     string            `json:"createdAt"`
	NameWithOwner string            `json:"nameWithOwner"`
}

type languageCount struct {
	name  string
	count int
}

func Red(s string) string {
	return "\x1b[31m" + s + "\x1b[m"
}

func Green(s string) string {
	return "\x1b[32m" + s + "\x1b[m"
}

func Yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[m"
}

func Blue(s string) string {
	return "\x1b[34m" + s + "\x1b[m"
}

func Gray(s string) string {
	return "\x1b[90m" + s + "\x1b[m"
}

func getAllRepos(org, language string, repoLimit int) ([]Repo, error) {

	fmt.Println(Yellow("Fetching repositories and analyzing languages..."))

	repolanguages, _, err := gh.Exec("repo", "list", org, "--limit", strconv.FormatInt(int64(repoLimit), 10), "--json", "nameWithOwner,languages,createdAt")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return []Repo{}, err
	}

	var repos []Repo
	jsonErr := json.Unmarshal(repolanguages.Bytes(), &repos)
	if jsonErr != nil {
		fmt.Printf("Error: %s\n", jsonErr.Error())
		return []Repo{}, jsonErr
	}

	// Sort repos by descending createdAt date
	sort.Slice(repos, func(i, j int) bool {
		iTime, _ := time.Parse(time.RFC3339, repos[i].CreatedAt)
		jTime, _ := time.Parse(time.RFC3339, repos[j].CreatedAt)
		return iTime.After(jTime)
	})

	fmt.Printf(Yellow("Analyzed %d repositories.\n"), len(repos))

	return repos, nil
}
