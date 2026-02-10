# General guidelines

- Use Go and the cobra library for command-line applications.
- Use [pterm](https://github.com/pterm/pterm) for terminal output.
- Keep reusable code in the `common.go` file.
- Always describe changes in a detailed plan before making them.
- Update the README file with any new features or changes.

# Naming Conventions
- Use PascalCase for component names, interfaces, and type aliases
- Use camelCase for variables, functions, and methods
- Prefix private class members with underscore (_)
- Use ALL_CAPS for constants

# Error Handling
- Use try/catch blocks for async operations
- Always log errors with contextual information
- Use logging to debug instead of print statements

# API preference

- When using the GitHub API, prefer using the REST API over the GraphQL API.
- An exception to this rule is when listing organizations for an enterprise. For this kind of task, always use the GraphQL API.
- If there is no confirmed API endpoint in the GitHub documentation, ask the user to provide one. Do NOT make assumptions about the endpoint.