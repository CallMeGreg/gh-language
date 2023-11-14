# Goal
## Subcommand tree (and command-specific flags)
- gh language count [org]

- gh language data [org]
   - --unit, -U (B/KB/GB/TB)

## Persistent flags
- --limit, -l (int) The maximum number of repositories to evaluate (default 100)
- --top, -t (int) Return the top N languages (default 10)
- --language, -L (string) The language to filter on
- --primary, -P (true/false) only look at the primary language of each repository (default false)
- --codeql, -C (all/compiled/interpreted)  only look for repos that contain a CodeQL compatible language (must be explicit)
- --report, -R (csv/json) format for a downloaded report
- --output, -O (string) location to save file


# v1.0.0
- gh language count [org] (--limit) (--top)

# v1.0.1
- gh language count [org] (--limit) (--top) (--language)
