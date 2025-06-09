# GitHub Language Analyzer

This is an extension to the `gh` command-line tool for analyzing the count of programming languages used in repositories across a GitHub enterprise or organization. It retrieves a list of repositories and their associated languages, and then aggregates the data to produce a report of language frequency.

> [!NOTE]
> If you are looking to compare your language frequency against public trends, you can access quarterly data from 2020 onward [here](https://innovationgraph.github.com/global-metrics/programming-languages) as part of GitHub's [Innovation Graph](https://innovationgraph.github.com/) project.

# Pre-requisites

1. Install the GitHub CLI: https://github.com/cli/cli#installation
2. Confirm that you are authenticated with an account that has access to the enterprise/org you would like to analyze:

```
gh auth status
```

Ensure that you have the necessary scopes. For example, if you are analyzing an organization, you need the `repo` scope and for enterprises you need the `read:enterprise` scope. You can add scopes by running:

```
gh auth login -s "repo,read:enterprise"
```

> [!IMPORTANT]
> Enterprise owners do not inherently have access to all of the repositories across their organizations. You must ensure that your account has the necessary permissions to access the repositories you want to analyze.

# Installation

To install this extension, run the following command:
```
gh extension install CallMeGreg/gh-language
```

# Usage

> [!TIP]
> Each command has default limits to prevent accidental excessive API usage. You can adjust these limits using the `--org-limit` and `--repo-limit` flags. To analyze all repositories in an organization or enterprise, set these flags to a very high number (e.g., `1000000`).

## Universal Flags

The following flags are available for all commands:
- `--org` or `--enterprise`: Specify the organization or enterprise to analyze. These flags are mutually exclusive, and one of them is required.
- `--org-limit`: Limit the number of organizations to analyze (default is 5).
- `--repo-limit`: Limit the number of repositories to analyze per organization (default is 10).
- `--top`: Return the top N languages (default is 10).
- `--language`: Filter results by a specific programming language (case-sensitive).
- `--codeql`: Restrict analysis to CodeQL-supported languages.

> [!NOTE]
> The `--top`, `--language`, and `--codeql` flags are mutually exclusive.

When the `--codeql` flag is set, the analysis will only include the following languages:
- C
- C++
- C#
- Go
- HTML
- Java
- Kotlin
- JavaScript
- Python
- Ruby
- Swift
- TypeScript
- Vue

## Count command

Display the count of each programming language used in repos across an enterprise or organization.
```
gh language count --enterprise YOUR_ENTERPRISE_SLUG
```

https://github.com/user-attachments/assets/27c0a12f-1643-4483-aeae-95aa61165879

## Trend command

Display the breakdown of programming languages used in repos across an enterprise or organization per year, based on the repo creation date.
```
gh language trend --enterprise YOUR_ENTERPRISE_SLUG
```

https://github.com/user-attachments/assets/33c1f4ac-57d9-4ed9-a696-0eb845cd6638

## Data command

Analyze languages by bytes of data, rather than count, across repositories in an enterprise or organization.
```
gh language data --enterprise YOUR_ENTERPRISE_SLUG
```

Specify the unit for displaying data with the `--unit` flag. Supported units are `bytes`, `kilobytes`, `megabytes`, and `gigabytes`. The default is `bytes`.:
```
gh language data --enterprise YOUR_ENTERPRISE_SLUG --unit megabytes
```

https://github.com/user-attachments/assets/435f1d81-2d56-4320-b3dd-7f8d6f2472bb

## Example Usage
Analyze the top 20 languages used across all repositories in an enterprise:
```
gh language count --enterprise YOUR_ENTERPRISE_SLUG --org-limit 1000000 --repo-limit 1000000 --top 20
```

Analyze the trend of Rust usage in repositories across an organization, limited to the first 100 repositories:
```
gh language trend --org YOUR_ORG_SLUG --repo-limit 100 --language Rust
```

Analyze the top 5 languages, based on data size, in megabytes, used across all repositories in an organization:
```
gh language data --org YOUR_ORG_SLUG --repo-limit 1000000 --top 5 --unit megabytes
```

Analyze all CodeQL-supported languages in an enterprise across all repositories:
```
gh language count --enterprise YOUR_ENTERPRISE_SLUG --org-limit 1000000 --repo-limit 1000000 --codeql
```

https://github.com/user-attachments/assets/bb8f9ccb-9f71-40b2-9dc4-8d1e34476afd

## Performance

The `count` and `trend` commands have been optimized to use GitHub's GraphQL API, which provides significant performance improvements over the REST API:

- **Reduced API calls**: GraphQL fetches repository and language data in a single request, eliminating the need for separate REST API calls per repository
- **Efficient pagination**: Uses cursor-based pagination with real-time progress tracking showing the current page being fetched
- **Better rate limiting**: GraphQL API has different rate limits than REST API, often allowing for faster data retrieval

The `data` command continues to use the REST API as it requires detailed byte-level language statistics that are only available through the REST endpoints.

## Help

For help, run:
```
gh language -h
```

``` 
Usage:
  language [command]

Available Commands:
  count       Analyze the count of programming languages used in repos across an organization
  data        Analyze language data by bytes
  help        Help about any command
  trend       Analyze the trend of programming languages used in repos across an organization over time

Flags:
      --codeql              Restrict analysis to CodeQL-supported languages
  -e, --enterprise string   Specify the enterprise
  -h, --help                help for language
  -l, --language string     The language to filter on (case-sensitive)
  -o, --org string          Specify the organization
      --org-limit int       The maximum number of organizations to evaluate for an enterprise (default 5)
      --repo-limit int      The maximum number of repositories to evaluate per organization (default 10)
  -t, --top int             Return the top N languages (ignored when a language is specified) (default 10)

Use "language [command] --help" for more information about a command.
```

# License
This tool is licensed under the MIT License. See the [LICENSE](https://github.com/CallMeGreg/gh-language/blob/main/LICENSE) file for details.