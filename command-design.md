# Subcommand tree (and command-specific flags)
- gh language

- gh language distribution [org]
    - --top, -T (positive integer)

- gh language data [org]
   - --unit, -U (B/KB/GB/TB)

# Persistent flags
- --primary, -P (true/false) only look at the primary language of each repository [default false]
- --codeql, -C (all/default/advanced)  only look for repos that contain a CodeQL compatible language [must be explicit]
- --lang, -L (string) only look for repos that contain this language
- --report, -R (csv/json) format for a downloaded report
- --output, -O (string) location to save file


# v1.0.0
- gh language distribution [org] (--top) (--primary)