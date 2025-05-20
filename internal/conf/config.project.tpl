# DO NOT MODIFY THE FOLLOWING LINES UNLESS YOU KNOW WHAT YOU ARE DOING
# These branch settings are only affecting the gitlab-flow command line tool while
# activating the gitlab-flow.

# The name of current project, this should only be configured while
# gitlab-flow could NOT extract the correct gitlab repository name from
# current directory.
# For example, `git clone remote_repository_location path/to/directory`
# directory is NOT same as the gitlab repository name, then you should
# configure the project name manually.
project_name = "{{.ProjectName}}"

# Global debug flag, if set to true, will output debug information
debug = {{.DebugMode}}

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