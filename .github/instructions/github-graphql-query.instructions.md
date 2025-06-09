---
applyTo: '**/*.go'
---


# Example GraphQL Query to fetch repository languages with pagination

```
query GetRepositories($org: String!, $repoCount: Int!, $cursor: String) {
  organization(login: $org) {
    repositories(first: $repoCount, after: $cursor) {
      nodes {
        name
        languages(first: 100) {
          nodes {
            name
          }
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}
```

# Exceeding the rate limit

If you exceed your primary rate limit, the response status will still be 200, but you will receive an error message, and the value of the x-ratelimit-remaining header will be 0. You should not retry your request until after the time specified by the x-ratelimit-reset header.

If you exceed a secondary rate limit, the response status will be 200 or 403, and you will receive an error message that indicates that you hit a secondary rate limit. If the retry-after response header is present, you should not retry your request until after that many seconds has elapsed. If the x-ratelimit-remaining header is 0, you should not retry your request until after the time, in UTC epoch seconds, specified by the x-ratelimit-reset header. Otherwise, wait for at least one minute before retrying. If your request continues to fail due to a secondary rate limit, wait for an exponentially increasing amount of time between retries, and throw an error after a specific number of retries.