# General guidelines

- Use Go and the cobra library for command-line applications.
- Use [pterm](https://github.com/pterm/pterm) for terminal output.
- Keep reusable code in the `common.go` file.
- Always describe changes in a detailed plan before making them.
- Update the README file with any new features or changes.

# API preference

- When using the GitHub API, prefer using the REST API over the GraphQL API.
- An exception to this rule is when listing organizations for an enterprise. For this kind of task, always use the GraphQL API.
- If there is no confirmed API endpoint, do not fabricate an endpoint.

# GitHub GraphQL API Rate Limits

- If you exceed your primary rate limit, the response status will still be 200, but you will receive an error message, and the value of the x-ratelimit-remaining header will be 0. You should not retry your request until after the time specified by the x-ratelimit-reset header.
- If you exceed a secondary rate limit, the response status will be 200 or 403, and you will receive an error message that indicates that you hit a secondary rate limit. If the retry-after response header is present, you should not retry your request until after that many seconds has elapsed. If the x-ratelimit-remaining header is 0, you should not retry your request until after the time, in UTC epoch seconds, specified by the x-ratelimit-reset header. Otherwise, wait for at least one minute before retrying. If your request continues to fail due to a secondary rate limit, wait for an exponentially increasing amount of time between retries, and throw an error after a specific number of retries.
- Continuing to make requests while you are rate limited may result in the banning of your integration.

# GitHub GraphQL API Enterprise endpoint schema

"""
An account to manage multiple organizations with consolidated policy and billing.
"""
type Enterprise implements Node {
  """
  The announcement banner set on this enterprise, if any. Only visible to members of the enterprise.
  """
  announcementBanner: AnnouncementBanner

  """
  A URL pointing to the enterprise's public avatar.
  """
  avatarUrl(
    """
    The size of the resulting square image.
    """
    size: Int
  ): URI!

  """
  The enterprise's billing email.
  """
  billingEmail: String

  """
  Enterprise billing information visible to enterprise billing managers.
  """
  billingInfo: EnterpriseBillingInfo

  """
  Identifies the date and time when the object was created.
  """
  createdAt: DateTime!

  """
  Identifies the primary key from the database.
  """
  databaseId: Int

  """
  The description of the enterprise.
  """
  description: String

  """
  The description of the enterprise as HTML.
  """
  descriptionHTML: HTML!

  """
  The Node ID of the Enterprise object
  """
  id: ID!

  """
  The location of the enterprise.
  """
  location: String

  """
  A list of users who are members of this enterprise.
  """
  members(
    """
    Returns the elements in the list that come after the specified cursor.
    """
    after: String

    """
    Returns the elements in the list that come before the specified cursor.
    """
    before: String

    """
    Only return members within the selected GitHub Enterprise deployment
    """
    deployment: EnterpriseUserDeployment

    """
    Returns the first _n_ elements from the list.
    """
    first: Int

    """
    Only return members with this two-factor authentication status. Does not
    include members who only have an account on a GitHub Enterprise Server instance.

    **Upcoming Change on 2025-04-01 UTC**
    **Description:** `hasTwoFactorEnabled` will be removed. Use `two_factor_method_security` instead.
    **Reason:** `has_two_factor_enabled` will be removed.
    """
    hasTwoFactorEnabled: Boolean = null

    """
    Returns the last _n_ elements from the list.
    """
    last: Int

    """
    Ordering options for members returned from the connection.
    """
    orderBy: EnterpriseMemberOrder = {field: LOGIN, direction: ASC}

    """
    Only return members within the organizations with these logins
    """
    organizationLogins: [String!]

    """
    The search string to look for.
    """
    query: String

    """
    The role of the user in the enterprise organization or server.
    """
    role: EnterpriseUserAccountMembershipRole

    """
    Only return members with this type of two-factor authentication method. Does
    not include members who only have an account on a GitHub Enterprise Server instance.
    """
    twoFactorMethodSecurity: TwoFactorCredentialSecurityType = null
  ): EnterpriseMemberConnection!

  """
  The name of the enterprise.
  """
  name: String!

  """
  A list of organizations that belong to this enterprise.
  """
  organizations(
    """
    Returns the elements in the list that come after the specified cursor.
    """
    after: String

    """
    Returns the elements in the list that come before the specified cursor.
    """
    before: String

    """
    Returns the first _n_ elements from the list.
    """
    first: Int

    """
    Returns the last _n_ elements from the list.
    """
    last: Int

    """
    Ordering options for organizations returned from the connection.
    """
    orderBy: OrganizationOrder = {field: LOGIN, direction: ASC}

    """
    The search string to look for.
    """
    query: String

    """
    The viewer's role in an organization.
    """
    viewerOrganizationRole: RoleInOrganization
  ): OrganizationConnection!

  """
  Enterprise information visible to enterprise owners or enterprise owners'
  personal access tokens (classic) with read:enterprise or admin:enterprise scope.
  """
  ownerInfo: EnterpriseOwnerInfo

  """
  The raw content of the enterprise README.
  """
  readme: String

  """
  The content of the enterprise README as HTML.
  """
  readmeHTML: HTML!

  """
  The HTTP path for this enterprise.
  """
  resourcePath: URI!

  """
  Returns a single ruleset from the current enterprise by ID.
  """
  ruleset(
    """
    The ID of the ruleset to be returned.
    """
    databaseId: Int!
  ): RepositoryRuleset

  """
  A list of rulesets for this enterprise.
  """
  rulesets(
    """
    Returns the elements in the list that come after the specified cursor.
    """
    after: String

    """
    Returns the elements in the list that come before the specified cursor.
    """
    before: String

    """
    Returns the first _n_ elements from the list.
    """
    first: Int

    """
    Returns the last _n_ elements from the list.
    """
    last: Int
  ): RepositoryRulesetConnection

  """
  The URL-friendly identifier for the enterprise.
  """
  slug: String!

  """
  Identifies the date and time when the object was last updated.
  """
  updatedAt: DateTime!

  """
  The HTTP URL for this enterprise.
  """
  url: URI!

  """
  A list of repositories that belong to users. Only available for enterprises with Enterprise Managed Users.
  """
  userNamespaceRepositories(
    """
    Returns the elements in the list that come after the specified cursor.
    """
    after: String

    """
    Returns the elements in the list that come before the specified cursor.
    """
    before: String

    """
    Returns the first _n_ elements from the list.
    """
    first: Int

    """
    Returns the last _n_ elements from the list.
    """
    last: Int

    """
    Ordering options for repositories returned from the connection.
    """
    orderBy: RepositoryOrder = {field: NAME, direction: ASC}

    """
    The search string to look for.
    """
    query: String
  ): UserNamespaceRepositoryConnection!

  """
  Is the current viewer an admin of this enterprise?
  """
  viewerIsAdmin: Boolean!

  """
  The URL of the enterprise website.
  """
  websiteUrl: URI
}

"""
The connection type for User.
"""
type EnterpriseAdministratorConnection {
  """
  A list of edges.
  """
  edges: [EnterpriseAdministratorEdge]

  """
  A list of nodes.
  """
  nodes: [User]

  """
  Information to aid in pagination.
  """
  pageInfo: PageInfo!

  """
  Identifies the total count of items in the connection.
  """
  totalCount: Int!
}