# gitlab-flow

[![Go Report Card](https://goreportcard.com/badge/github.com/yeqown/gitlab-flow)](https://goreportcard.com/report/github.com/yeqown/gitlab-flow) [![go.de
│ v reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/yeqown/gitlab-flow)

[English Document](./README.md)

这是一个用于管理 `gitlab` 开发流程的命令行工具。它和 `git-flow` 的区别在于 `gitlab-flow` 会操作远程资源，比如说： 里程碑，issue, 合并请求和
分支。更重要的是，如果是在团队中使用 `gitlab-flow` 可以帮助你快速同步其他团队成员的迭代信息。

<img src="./assets/intro.svg" width="100%"/>

### 安装

因为仓库没有提供预编译的可执行文件，所以需要你自己安装。

```shell
APP_ID=YOUR_GITLAB_APP_ID \
APP_SECRET=YOUR_GITLAB_APP_SECRET \
make build
```

### 初始化

```shell
gitlab-flow [-c, --conf `path/to/confpath/`] init
# 注意：全局选项必须置于子命令之前；
# -c 参数只需要配置路径即可，而不需要指定文件；
```

#### Gitlab 授权

> !!! 在你初始化 gitlab-flow 客户端之前，你必须使用【特殊编译】的 gitlab-flow 可执行文件。
> 特殊编译是指：将在你的 gitlab 服务器上创建的 gitlab-flow 应用的应用ID和应用密钥编译到可执行文件中。

在初始化命令执行后，gitlab-flow会通过交互式命令行采集你的本地配置，并在配置采集完成后自动运行授权程序。

> Gitlab 服务地址: The domain of your gitlab server. such as https://git.example.com
>
> Gitlab API 域名: gitlab服务的 API 端点. 譬如: https://git.example.com/api/v4/。
> 你可以在这里找到:
[https://git.example.com/help/api/README.md](https://git.example.com/help/api/README.md).
这个页面提供一些示范API，你可以从中找到gitlab服务的API域名。

### CLI Help  

```shell
$ flow -h
NAME:
   gitlab-flow - 命令行工具

USAGE:
   flows [全局选项] 命令 [命令选项] [参数...]

VERSION:
   v1.6.2

DESCRIPTION:
   用于管理 gitlab 中的 Feature/Milestone/Issue/MergeRequest 资源.

AUTHOR:
   yeqown <yeqown@gmail.com>

COMMANDS:
   help, h  Shows a list of commands or help for one command
   dash:
     dash  概览命令
   flow:
     feature  迭代开发命令
     hotfix   热修复命令
   init:
     init  初始化命令

GLOBAL OPTIONS:
   --conf path/to/file, -c path/to/file  指定配置路径 (default: ~/.gitlab-flow)
   --cwd path/to/file                         指定工作路径 (default: 当前路径)
   --debug                                    调试模式 (default: false)
   --project projectName, -p projectName      指定项目名 (default: 当前目录名)
   --force-remote                             是否强制从远程匹配项目, 而不是从本地 (default: false)
   --web                                      是否打开浏览器 (default: false)
   --help, -h                                 帮助信息 (default: false)
   --version, -v                              版本号 (default: false)
```

### [文档](./docs/zh-CN/README.md)

<img align="center" src="./assets/gitlab-flow-branch.png">
