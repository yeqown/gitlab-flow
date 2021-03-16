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
go install github.com/yeqown/gitlab-flow/cmd/gitlab-flow

# 如果不能通过上述命令直接安装，你可以尝试如下的命令，然后再执行安装命令。
go get -u github.com/yeqown/gitlab-flow
```

### 初始化

```shell
gitlab-flow [-c, --conf_path `path/to/confpath/`] init -s "YOUR_GITLAB_ACCESS_TOKEN" -d "YOUR_GITLAB_API_HOST"
# 注意：全局选项必须置于子命令之前；
# -c 参数只需要配置路径即可，而不需要指定文件；
```

#### Gitlab 授权

你可以在如下的地址去创建: 

[https://git.example.com/profile/personal_access_tokens](https://git.example.com/profile/personal_access_tokens).

你需要选中 `api, read_user, read_repository, read_registry` 等作用域.

#### Gitlab API 域名

可以在这里找到:

[https://git.example.com/help/api/README.md](https://git.example.com/help/api/README.md).

这个页面提供一些示范API，你可以从中找到gitlab服务的API域名.

### CLI Help  

```shell
$ flow -h
NAME:
   gitlab-flow - 命令行工具

USAGE:
   flow2 [全局选项] 命令 [命令选项] [参数...]

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
   --conf_path path/to/file, -c path/to/file  指定配置路径 (default: ~/.gitlab-flow)
   --cwd path/to/file                         指定工作路径 (default: 当前路径)
   --debug                                    调试模式 (default: false)
   --project projectName, -p projectName      指定项目名 (default: 当前目录名)
   --force-remote                             是否强制从远程匹配项目, 而不是从本地 (default: false)
   --web                                      是否打开浏览器 (default: false)
   --help, -h                                 帮助信息 (default: false)
   --version, -v                              版本号 (default: false)
```

### 文档

<img align="center" src="./assets/gitlab-flow-branch.png">

查看更多说明: [文档](./docs/zh-CN/README.md).
