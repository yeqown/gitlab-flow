## Configuration

The gitlab-flow provides hierarchical configurations:

- Global config file: `~/.gitlab-flow/config.toml`
- Project config file: `path/to/project-roo/.gitlab-flow/config.toml`

gitlab-flow provides a subcommand to manage the configuration file, you can use
`flow config --help` to see the usage.

```sh
$ flow2 config --help
NAME:
   gitlab-flow config - show current configuration

USAGE:
   gitlab-flow config command [command options] [arguments...]

COMMANDS:
   init     initialize configuration gitlab-flow, generate default config file and sqlite DB
   show     show current configuration
   edit     edit current configuration
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --global, -g  show global configuration (default: false)
   --help, -h    show help (default: false)
```

### Global Configuration

gitlab-flow provides an initialization command to initialize the global configuration file.

```bash
flow config --global init
```

```toml
# Global debug flag, if set to true, will output debug information
debug = false

# The gitlab API URL which is used to interact with gitlab.
gitlab_api_url = "https://git.xxxx.com/api/v4"

# The gitlab host URL which is used to interact with gitlab, for example
# 1. open gitlab project page
# 2. OAuth2 authorization
# 3. etc.
gitlab_host = "https://git.xxxx.com"

# A flag which controls gitlab-flow to open browser automatically or not.
# If set to true, gitlab-flow always open browser automatically.
open_browser = false

# The branch settings controls the branch name which gitlab-flow would access,
# generate, and use. for example, while gitlab-flow is creating a feature branch,
# it will use the branch name (FeatureBranchPrefix + feature name), and checkout
# the branch from the Master.
[branch]
  master = "main"
  dev = "dev"
  test = "test"
  # NOTICE: DO NOT CHANGE prefixes of the branch settings unless you're re-initializing
  # the gitlab-flow.
  feature_branch_prefix = "version/"
  hotfix_branch_prefix = "hotfix/"
  conflict_resolve_branch_prefix = "conflict-resolve/"
  issue_branch_prefix = "issue/"

# OAuth2 settings, which stores the access token and refresh token for gitlab-flow
# to access gitlab API.
# And the scope is the OAuth2 scope which gitlab-flow would request and callback URI.
[oauth]
  access_token = "xxxx"
  refresh_token = "xxxxx"
  app_id = "xxxxx"
  app_secret = "xxxx"
  # DO NOT MODIFY THE FOLLOWING LINES UNLESS YOU KNOW WHAT YOU ARE DOING
  scopes = "api read_user read_repository"
  callback_host = "localhost:2333"
  # The mode indicates the OAuth2 mode, 1 for authorization automatically, 2 for manual authorization which should
  # be used only headless(this means current system could not open browser, e.g. linux server) environment.
  mode = 1
```

### Project Configuration

gitlab-flow provides a initialization command to initialize the project configuration file.

```bash
flow config init
```

```toml
# DO NOT MODIFY THE FOLLOWING LINES UNLESS YOU KNOW WHAT YOU ARE DOING
# These branch settings are only affecting the gitlab-flow command line tool while
# activating the gitlab-flow.

# The name of current project, this should only be configured while
# gitlab-flow could NOT extract the correct gitlab repository name from
# current directory.
# For example, `git clone remote_repository_location path/to/directory`
# directory is NOT same as the gitlab repository name, then you should
# configure the project name manually.
project_name = "xxx"

# Global debug flag, if set to true, will output debug information
debug = false

# A flag which controls gitlab-flow to open browser automatically or not.
# If set to true, gitlab-flow always open browser automatically.
open_browser = false

# The branch settings controls the branch name which gitlab-flow would access,
# generate, and use. for example, while gitlab-flow is creating a feature branch,
# it will use the branch name (FeatureBranchPrefix + feature name), and checkout
# the branch from the Master.
[branch]
  master = "main"
  dev = "dev"
  test = "test"
  # NOTICE: DO NOT CHANGE prefixes of the branch settings unless you're re-initializing
  # the gitlab-flow.
  feature_branch_prefix = "version/"
  hotfix_branch_prefix = "hotfix/"
  conflict_resolve_branch_prefix = "conflict-resolve/"
  issue_branch_prefix = "issue/"
```