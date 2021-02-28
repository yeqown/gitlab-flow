## 配置文件

```toml
# access_token 用于授权 gitlab-flow 使用 GITLAB 开放接口。
access_token = "YOUR_ACCESS_TOKEN"

# gitlab_api_url 是 GITLAB 的开放接口的地址。 
gitlab_api_url = "YOUR_GITLAB_API_HOST"

# debug 调试模式开关，主要用于开发，以及在使用过程中遇到问题时帮助定位问题。
# 在全局命令参数中 可以通过 --debug 来覆盖配置文件的开关。
debug = false

# open_browser 表明是否需要自动打开浏览器，以定位到某些资源（MR，ISSUE的网络地址）。
# true 表示开启，false 表示不开启。
open_browser = true
```