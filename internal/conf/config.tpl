# Global debug flag, if set to true, will output debug information
debug = {{.DebugMode}}

# The gitlab API URL which is used to interact with gitlab.
gitlab_api_url = "{{.GitlabAPIURL}}"

# The gitlab host URL which is used to interact with gitlab, for example
# 1. open gitlab project page
# 2. OAuth2 authorization
# 3. etc.
gitlab_host = "{{.GitlabHost}}"

# A flag which controls gitlab-flow to open browser automatically or not.
# If set to true, gitlab-flow always open browser automatically.
open_browser = {{.OpenBrowser}}

# The branch settings controls the branch name which gitlab-flow would access,
# generate, and use. for example, while gitlab-flow is creating a feature branch,
# it will use the branch name (FeatureBranchPrefix + feature name), and checkout
# the branch from the Master.
[branch]
  master = "{{.Branch.Master}}"
  dev = "{{.Branch.Dev}}"
  test = "{{.Branch.Test}}"
  # NOTICE: DO NOT CHANGE prefixes of the branch settings unless you're re-initializing
  # the gitlab-flow.
  feature_branch_prefix = "{{.Branch.FeatureBranchPrefix}}"
  hotfix_branch_prefix = "{{.Branch.HotfixBranchPrefix}}"
  conflict_resolve_branch_prefix = "{{.Branch.ConflictResolveBranchPrefix}}"
  issue_branch_prefix = "{{.Branch.IssueBranchPrefix}}"

# OAuth2 settings, which stores the access token and refresh token for gitlab-flow
# to access gitlab API.
# And the scope is the OAuth2 scope which gitlab-flow would request and callback URI.
[oauth]
  access_token = "{{.OAuth2.AccessToken}}"
  refresh_token = "{{.OAuth2.RefreshToken}}"
  # DO NOT MODIFY THE FOLLOWING LINES UNLESS YOU KNOW WHAT YOU ARE DOING
  scopes = "api read_user read_repository"
  callback_host = "{{.OAuth2.CallbackHost}}"