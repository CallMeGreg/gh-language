---
applyTo: '**/*.go'
---
Coding standards, domain knowledge, and preferences that AI should follow.

# Example Client API

package main

import (
	"fmt"
	"log"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
)

func main() {
	// These examples assume `gh` is installed and has been authenticated.

	// Use an API client to retrieve repository tags.
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}
	response := []struct{
		Name string
	}{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}

# Example REST Client with pagination

// Get releases from cli/cli repository using REST API with paginated results.
func ExampleRESTClient_pagination() {
	var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)
	findNextPage := func(response *http.Response) (string, bool) {
		for _, m := range linkRE.FindAllStringSubmatch(response.Header.Get("Link"), -1) {
			if len(m) > 2 && m[2] == "next" {
				return m[1], true
			}
		}
		return "", false
	}
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}
	requestPath := "repos/cli/cli/releases"
	page := 1
	for {
		response, err := client.Request(http.MethodGet, requestPath, nil)
		if err != nil {
			log.Fatal(err)
		}
		data := []struct{ Name string }{}
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Page: %d\n", page)
		fmt.Println(data)
		var hasNextPage bool
		if requestPath, hasNextPage = findNextPage(response); !hasNextPage {
			break
		}
		page++
	}
}