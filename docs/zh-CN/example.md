## 示范

此节提供了一些示例来说明，如何使用`gitlab-flow`来管理开发中的分支和相关资源。

### 0. 全局参数

```sh
flow [-c, --conf] [--debug] [--web] [-p, --project] [--force-remote] SUB_COMMAND [options]
# (可选) -c, --conf 配置文件路径。
# (可选) --debug 调试模式。
# (可选) --web 是否自动打开浏览器，来访问某些资源（MR/ISSUE的网络地址）。
# (可选) -p, --project 项目名，默认使用当前目录名。
# (可选) --force-remote 是否强制从远程匹配项目，而不是仅从本地数据库（用于项目重名时，更新本地项目列表）。

# 举例:
flow -c ~/.gitlab-flow --debug init ...
# 上述命令的含义是：指定了 `~/.gitlab-flow` 作为gitlab-flow配置路径，并在调试模式下初始化。
```

### 0.1 迭代命令选项

```sh
flow [global options] feature [feature options] SUB_COMMAND [options]
# (可选) -f, --feature-branch-name 指定迭代分支名，默认当分支名
# (可选) --force-create-mr 强制从远程创建MR，而不是先从本地查询（用于某一个分支再次合并到特定分支）.

# 举例:
flow --web feature -f featureBranchName --force-create-mr debug
# 上述命令的含义是：强制创建一个 featureBranchName 分支合并到开发分支的 MR，并自动打开浏览器。
```


### 1. 开启一个迭代分支

```sh
flow feature [-f, --feature_branch_name featureBranchName] open name description
# (必填) feature-name 用于迭代分支名和里程碑名.
# (必填) feature-description 用于里程碑的描述信息.
#
# 执行结果:
# feature/feature-name 分支会被创建；里程碑 featureBranchName 会被创建.
```

### 2. 完成一次迭代

```sh
flow feature [-f, --feature_branch_name featureBranchName] release
# (可选) -f, --feature_branch_name 指定featureBranchName迭代已经完成，创建该迭代分支到master的MR，
# 如果没有设置，gitlab-flow 会使用当前分支，作为迭代分支。
```

### 3. 开启一次迭代中的小特性（issue分支）

```sh
flow feature [-f, --feature_branch_name featureBranchName] open-issue issue-title issue-description
# (可选) -f, --feature_branch_name featureBranchName 指明此特性属于哪一次迭代，
# 如果没有设置，gitlab-flow 会使用当前分支，作为迭代分支。
# (必填) issue-name 指明issue的名字
# (必填) issue-description 指明issue的描述信息.
```

### 4. 完成一次迭代中的小特性

```sh
flow feature [-f, --feature_branch_name featureBranchName] close-issue [-i, --issue_branch_name issueBranchName]
# (可选) -i, --issue_branch_name issueBranchName, 指明哪一个issue分支需要被完成，
# 如果没有设置，会默认使用当前分支。
# (可选) -f, --feature_branch_name featureBranchName, 指明此次小特性属于那一次迭代，
# 如果没有设置，会根据issue分支确定
```

### 5. 开启一次热修复

```sh
flow hotfix open hotfix-name hotfix-description
# (必填) hotfix-name 指明hotfix关联的issue的标题和hotfix的分支名
# (必填) hotfix-description hotfix关联的issue的描述信息
```
### 6. 完成热修复

```sh
flow hotfix close [-b, --branch_name hotfixBranchName]
# (可选) -b, --branch_name 指明哪一个热修复分支需要被完成。
```

### 7. 从远程迭代数据到本地

```sh
flow feature sync [-m, --milestone_id milestoneId] [-i, --interact]
# (可选) -m, --milestone_id 里程碑ID，指明哪次迭代的相关资源需要同步
# (可选) -i, --interact, 交互模式，推荐使用
#
# NOTE: 至少选择一种同步方式，如果两种都设置了那么会优先使用 milestone_id。
```

### 8. 解决冲突流程（master或者其他分支）

```sh
flow --web feature [-f, --feature-branch-name featureBranchName] resolve-conflict \ 
	[-t, --target_branch targetBranchName]
# (可选) -f, --feature-branch-name 指明哪一次迭代需要解决冲突，默认是当前分支
# (可选) -t, --target_branch targetBranchName, 指明目标分支，默认是master分支


# 注意：此命令会在本地执行 `git merge --no-ff $featureBranchName`，需要保证本地的目标分支代码是最新的。
# 1. 从目标分支创建一个 resolve-conflict/feature-branch
# 2. 创建一个 resolve-conflict/feature-branch 到目标分支的 MR
# 3. 本地 切换到 resolve-conflict/feature-branch 执行 `git merge --no-ff $featureBranchName`
```
