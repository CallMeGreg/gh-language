# GitHub Language Analyzer

This is an extension to the `gh` command-line tool for analyzing the count of programming languages used in repositories across a GitHub enterprise or organization. It retrieves a list of repositories and their associated languages, and then aggregates the data to produce a report of language frequency.

> [!NOTE]
> If you are looking to compare your language frequency against public trends, you can access quarterly data from 2020 onward [here](https://innovationgraph.github.com/global-metrics/programming-languages) as part of GitHub's [Innovation Graph](https://innovationgraph.github.com/) project.

# Pre-requisites

1. Install the GitHub CLI: https://github.com/cli/cli#installation
2. Confirm that you are authenticated with an account that has access to the org you would like to analyze:

```
gh auth status
```

Ensure that you have the necessary scopes. For example, if you are analyzing an organization, you need `repo` scope and for enterprises you need the `read:enterprise` scope. You can add scopes by running:

```
gh auth login -s "repo,read:enterprise"
```

# Installation

To install this extension, run the following command:
```
gh extension install CallMeGreg/gh-language
```

# Usage

## Count command

Display the count of each programming language used in repos across an enterprise or organization.
```
gh language count --enterprise YOUR_ENTERPRISE_SLUG
```

Optionally specify the organization limit (`--org-limit`), repo limit (`--repo-limit`) and the number of languages to return (`--top`)
```
gh language count --enterprise YOUR_ENTERPRISE_SLUG --org-limit 5 --repo-limit 100 --top 10
```

Optionally filter by a specific language (`--language`)
```
gh language count YOUR_ORG_NAME --language Java
```
> [!IMPORTANT]
> The `--language` flag values are case-sensitive.

## Trend command

Display the breakdown of programming languages used in repos across an enterprise or organization per year, based on the repo creation date.
```
gh language trend --org YOUR_ORG_NAME
```

## Data command

Analyze language data by bytes, rather than count, across repositories in an enterprise or organization.
```
gh language data --org YOUR_ORG_NAME
```

Specify the unit for displaying data (`--unit`):
```
gh language data --org YOUR_ORG_NAME --unit megabytes
```

## Flags

The following flags are available for all commands:
- `--org` or `--enterprise`: Specify the organization or enterprise to analyze. These flags are mutually exclusive, and one of them is required.
- `--org-limit`: Limit the number of organizations to analyze (default is 5).
- `--repo-limit`: Limit the number of repositories to analyze per organization (default is 10).
- `--top`: Return the top N languages (default is 10). This flag is ignored when a specific language is specified.
- `--language`: Filter results by a specific programming language (case-sensitive).

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
  data        Analyze language data by bytes for a specific repository
  help        Help about any command
  trend       Analyze the trend of programming languages used in repos across an organization over time

Flags:
  -h, --help              help for language
  -L, --language string   The language to filter on
  -l, --limit int         The maximum number of repositories to evaluate (default 100)
  -t, --top int           Return the top N languages (ignored when a language is specified) (default 10)

Use "language [command] --help" for more information about a command.
```

# License
This tool is licensed under the MIT License. See the [LICENSE](https://github.com/CallMeGreg/gh-language/blob/main/LICENSE) file for details.
