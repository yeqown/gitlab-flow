## Examples

There are some examples of how to using `gitlab-flow` to help manage 
development resources.

### 0. Global flags

```shell
flow [-c, --conf_path] [--debug] [--web] [-p, --project] SUB_COMMAND [options]
# (OPTIONAL) -c, --conf_path path/to/config_file.
# (OPTIONAL) --debug verbose mode.
# (OPTIONAL) --web open web browser of resource url automatically.
# (OPTIONAL) -p, --project projectName of current working directory.

# example:
flow -c ~/.gitlab-flow --debug init ...
# means initialize gitlab-flow config_file in specified 
# directory `~/.gitlab-flow` and logs will be verbose
```

### 1. Start a milestone feature.

```shell
flow feature open name description
# (REQUIRED) feature-name will be used to create milestone as title too.
# (REQUIRED) feature-description will be to create milestone as description too.
#
# RESULT:
# feature/feature-name is your feature branch name.
```

### 2. Finish a milestone feature.

```shell
flow feature release [-f, --feature_branch_name featureBranchName]
# (OPTIONAL) -f, --feature_branch_name featureBranchName, if it is not set,
# current branch name will be used.
```

### 3. Start an issue from a feature.

```shell
flow feature open-issue [-f, --feature_branch_name featureBranchName] issue-title issue-description
# (OPTIONAL) -f, --feature_branch_name featureBranchName, if it is not set,
# will find feature branch name relative to issue branch name.
# (REQUIRED) issue-name will be used to create issue as title too.
# (REQUIRED) issue-description will be to create issue as description too.
```

### 4. Finish an issue from a feature.

```shell
flow feature close-issue [-i, --issue_branch_name issueBranchName] [-f, --feature_branch_name featureBranchName] 
# (OPTIONAL) -i, --issue_branch_name issueBranchName, if it is not set,
# current branch name will be used.
# (OPTIONAL) -f, --feature_branch_name featureBranchName, if it is not set,
# will find feature branch name relative to issue branch name.
```

### 5. Start a hotfix.

```shell
flow hotfix open hotfix-name hotfix-description
# (REQUIRED) hotfix-name will be used to create issue as title too.
# (REQUIRED) hotfix-description will be to create issue as description too.
#
# RESULT:
# hotfix/hotfix-name is your feature branch name.
```
### 6. Finish a hotfix.

```shell
flow hotfix close [-b, --branch_name hotfixBranchName]
# (OPTIONAL) -b, --branch_name hotfixBranchName, if it is not set,
# current branch name will be used.
```