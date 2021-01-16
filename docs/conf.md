## Configuration

```toml
# access_token is the access_token to access gitlab API.
access_token = "YOUR_ACCESS_TOKEN"

# gitlab_api_url is the base url where gitlab-flow client requests resources from. 
gitlab_api_url = "YOUR_GITLAB_API_HOST"

# debug mode to print more information when you got problem
# you can also use --debug in global options to override debug switch in config file.
debug = false

# open_browser to indicate should gitlab-flow open resources' web url in your browser.
# true means always open, false means never open. 
open_browser = true
```