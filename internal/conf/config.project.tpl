# DO NOT MODIFY THE FOLLOWING LINES UNLESS YOU KNOW WHAT YOU ARE DOING
# These branch settings are only affecting the gitlab-flow command line tool while
# activating the gitlab-flow.
#
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