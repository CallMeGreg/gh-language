# GitHub Language Analyzer
This is an extension to the `gh` command-line tool for analyzing the count of programming languages used in repositories across a GitHub organization. It retrieves a list of repositories and their associated languages, and then aggregates the data to produce a report of language frequency.

# Pre-requisites
1. Install the GitHub CLI: https://github.com/cli/cli#installation
2. Confirm that you are authenticated with an account that has access to the org you would like to analyze:

```
gh auth status
```

# Installation
To install this extension, run the following command:
```
gh extension install CallMeGreg/gh-language
```

# Usage
Simplest example:
```
gh language count YOUR_ORG_NAME
```

Optionally specify the repo limit (`--limit`) and/or the number of languages to return (`--top`)
```
gh language count YOUR_ORG_NAME --limit 1000 --top 20
```

For help, run:
```
gh language -h
```

``` 
Usage:
  language [command]

Available Commands:
  count       Analyze the count of programming languages used in repos across an organization
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
