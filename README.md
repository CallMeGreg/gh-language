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
- `--org` or `--enterprise` (`-e`): Specify the organization or enterprise slug to analyze. These flags are mutually exclusive, and one of them is required.
- `--org-limit`: Limit the number of organizations to analyze (default is 5).
- `--repo-limit`: Limit the number of repositories to analyze per organization (default is 10).
- `--top`: Return the top N languages (default is 10).
- `--language`: Filter results by one or more programming languages, specified as a comma-separated list (case-sensitive).
- `--codeql`: Restrict analysis to CodeQL-supported languages.
- `--github-enterprise-server-url` (`-u`): GitHub Enterprise Server URL (e.g., github.company.com) (default is "github.com").

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
gh language count --org microsoft
```

![count](demo/count.gif)

When the `--codeql` flag is set, the analysis will filter for only CodeQL-supported languages, and also display the number of unique repositories that include at least one CodeQL-supported language:
```
gh language count --org microsoft --repo-limit 300 --codeql
```

![count-codeql](demo/count-codeql.gif)

## Trend command

Display the breakdown of programming languages used in repos across an enterprise or organization per year, based on the repo creation date. The output includes:

- **Line Graph** — A multi-series line chart showing how each language's adoption has changed over time.
- **Year-by-Year Breakdown** — Detailed per-year tables with trend direction and year-over-year deltas compared to the prior year.

```
gh language trend --org microsoft
```

![trend](demo/trend.gif)

The `trend` command also supports optional year range filtering:
- `--min-year`: Only include years greater than or equal to this value.
- `--max-year`: Only include years less than or equal to this value (defaults to last year).

```
gh language trend --org microsoft --repo-limit 500 --max-year 2015
```

![trend-filtered](demo/trend-filtered.gif)

## Data command

Analyze languages by bytes of data, rather than count, across repositories in an enterprise or organization.
```
gh language data --org microsoft
```

![data](demo/data.gif)

Specify the unit for displaying data with the `--unit` flag. Supported units are `bytes`, `kilobytes`, `megabytes`, and `gigabytes`. The default is `bytes`:
```
gh language data --org microsoft --unit megabytes
```

## Targeting a GitHub Enterprise Server instance

To target a GitHub Enterprise Server instance, use the `--github-enterprise-server-url` (`-u`) flag with any command. For example, to count languages across all repositories in an enterprise:
```
gh language count --enterprise github -u callmegreg-db6woz.ghe-test.net --org-limit 1000000 --repo-limit 1000000
```

![ghes](demo/ghes.gif)

## Performance

The `count` and `trend` commands have been optimized to use GitHub's GraphQL API, which provides significant performance improvements over the REST API. These commands are expected to run ~100x faster than `data`.

The `data` command continues to use the REST API as it requires detailed byte-level language statistics that are only available through the REST endpoint.

## Help

For help, run:
```
gh language -h
```

``` 
Usage:
  language [command]

Available Commands:
  count       Analyze the count of programming languages used in repos across an enterprise or organization
  data        Analyze the programming languages used in repos across an enterprise or organization based on bytes of data
  help        Help about any command
  trend       Analyze the trend of programming languages used in repos across an enterprise or organization over time

Flags:
      --codeql                                Restrict analysis to CodeQL-supported languages (mutually exclusive with --language, --top)
  -e, --enterprise string                     GitHub Enterprise slug (e.g., github)
  -u, --github-enterprise-server-url string   GitHub Enterprise Server URL (e.g., github.company.com) (default "github.com")
  -h, --help                                  help for language
  -l, --language string                       A comma-separated list of languages to filter on (case-sensitive, mutually exclusive with --codeql, --top)
  -o, --org string                            Specify the organization
      --org-limit int                         The maximum number of organizations to analyze for an enterprise (default 5)
      --repo-limit int                        The maximum number of repositories to analyze per organization (default 10)
  -t, --top int                               Return the top N languages (mutually exclusive with --language, --codeql) (default 10)

Use "gh language [command] --help" for more information about a command.
```

# License
This tool is licensed under the MIT License. See the [LICENSE](https://github.com/CallMeGreg/gh-language/blob/main/LICENSE) file for details.