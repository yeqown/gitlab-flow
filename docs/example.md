## Examples

There are some examples of how to use `gitlab-flow` to help manage 
development resources.

### 0. Global flags

```sh
flow [global options] command [command options] [arguments...]
GLOBAL OPTIONS:
   --debug                                verbose mode (default: false)
   --force-remote                         query project from remote not from local. This should be used when project name is duplicated, and could not found from local. (default: false)
   --web                                  open web browser automatically or not (default: false)
   --cwd path/to/file                     choose which path/to/file to load
   --project projectName, -p projectName  input projectName to locate which project should be operate.
   --help, -h                             show help (default: false)
   --version, -v                          print the version (default: false)
```

### 0.1 Feature flags

```sh
flow [global options] feature [feature options] SUB_COMMAND [options]
# (OPTIONAL) -f, --feature-branch-name feature branch name.
# (OPTIONAL) --force-create-mr force to create merge request in remote not query from local firstly.

# example:
flow --web feature -f featureBranchName --force-create-mr debug
# means flow would create merge request directly not query from local firstly 
# with specified branch name `featureBranchName`.
```


### 1. Start a feature development.

```sh
flow feature [-f, --feature_branch_name featureBranchName] open name description
# (REQUIRED) feature-name will be used to create milestone as title too.
# (REQUIRED) feature-description will be to create milestone as description too.
#
# RESULT:
# feature/feature-name is your feature branch name.
```

### 2. Finish a milestone feature.

```sh
flow feature [-f, --feature_branch_name featureBranchName] release
# (OPTIONAL) -f, --feature_branch_name featureBranchName, if it is not set,
# current branch name will be used.
```

### 3. Start an issue from a feature.

```sh
flow feature [-f, --feature_branch_name featureBranchName] open-issue issue-title issue-description
# (OPTIONAL) -f, --feature_branch_name featureBranchName, if it is not set,
# will find feature branch name relative to issue branch name.
# (REQUIRED) issue-name will be used to create issue as title too.
# (REQUIRED) issue-description will be to create issue as description too.
```

### 4. Finish an issue from a feature.

```sh
flow feature [-f, --feature_branch_name featureBranchName] close-issue [-i, --issue_branch_name issueBranchName]
# (OPTIONAL) -i, --issue_branch_name issueBranchName, if it is not set,
# current branch name will be used.
# (OPTIONAL) -f, --feature_branch_name featureBranchName, if it is not set,
# will find feature branch name relative to issue branch name.
```

### 5. Start a hotfix.

```sh
flow hotfix open hotfix-name hotfix-description
# (REQUIRED) hotfix-name will be used to create issue as title too.
# (REQUIRED) hotfix-description will be to create issue as description too.
#
# RESULT:
# hotfix/hotfix-name is your feature branch name.
```
### 6. Finish a hotfix.

```sh
flow hotfix close [-b, --branch_name hotfixBranchName]
# (OPTIONAL) -b, --branch_name hotfixBranchName, if it is not set,
# current branch name will be used.
```

### 7. Synchronize development

```sh
flow feature sync [-m, --milestone_id milestoneId] [-i, --interact]
# (OPTIONAL) -m, --milestone_id milestoneId input milestoneId 
# which you want to synchronize.
# (OPTIONAL) -i, --interact, if you don't know milestoneId, 
# then choose one milestone reciprocally.
#
# NOTE: at least one way should be chosen. if both of them are valued, 
# milestoneId has higher priority.
```

### 8. Resolve conflict between feature branch and master (or other target branch)

```sh
flow --web feature [-f, --feature-branch-name featureBranchName] resolve-conflict \ 
	[-t, --target_branch targetBranchName]
# (OPTIONAL) -f, --feature-branch-name which feature branch you wanna resolve conflicts.
# (OPTIONAL) -t, --target_branch targetBranchName, default is master.

# Notice: this command would execute `git merge --no-ff $featureBranchName`, of course it would make sure that
# current branch is your target branch and it has the latest codes. It also create a merge request between
# `conflict-resolve/feature-branch` into target branch.
```