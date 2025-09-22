# Flow3

基于 glab 的轻量级 GitLab 工作流工具（第三代），提供简化的开发流程管理。

## 特性

- 🚀 基于官方 glab 工具，稳定可靠
- 📦 零配置，开箱即用
- 🔄 标准化的 feature/hotfix 工作流
- 🎯 Issue 驱动开发支持
- 🤝 自动创建 MR 和关联 Issue

## 安装

### 前置依赖

```bash
# 安装 glab
brew install glab

# 或者从官网下载
# https://gitlab.com/gitlab-org/cli
```

### 安装 flow3

```bash
# 克隆或下载 flow3
./install-flow-cli.sh

# 初始化配置
flow3 init
```

## 使用方法

### Feature 开发流程

```bash
# 开始新功能开发
flow3 feature start user-authentication

# 开始功能开发并关联到 milestone
flow3 feature start user-auth 5

# 完成功能开发（自动创建 MR）
flow3 feature finish

# 完成功能开发到指定分支
flow3 feature finish develop
```

### Issue 驱动开发

```bash
# 开始处理 Issue #123
flow-cli issue start 123

# 完成 Issue（自动创建 MR 并关闭 Issue）
flow-cli issue finish
```

### 热修复流程

```bash
# 开始热修复
flow-cli hotfix start critical-security-fix

# 完成热修复
flow-cli hotfix finish
```

### 同步和查看

```bash
# 查看活跃的 milestones
flow-cli sync milestones

# 查看当前配置
flow-cli config
```

## 工作流说明

### Feature 流程
1. 从 main 分支创建 `feature/name` 分支
2. 可选择关联到指定 milestone
3. 完成后自动推送并创建 MR
4. MR 合并后自动删除源分支

### Issue 流程
1. 从 main 分支创建 `feature/issue-{id}` 分支
2. 完成后创建 MR 并自动关联 Issue
3. MR 标题包含 "Resolve #issue-id"

### Hotfix 流程
1. 从 main 分支创建 `hotfix/name` 分支
2. 完成后直接合并到 main 分支

## 配置

Flow3 使用内置默认配置，可通过项目根目录的 `.flow3/config` 文件覆盖：

```bash
# 初始化项目配置
flow3 init

# 编辑 .flow3/config
DEFAULT_TARGET_BRANCH=develop
FEATURE_PREFIX=feat/
HOTFIX_PREFIX=fix/
AUTO_CREATE_MR=false
```

默认配置：
- `DEFAULT_TARGET_BRANCH=main`
- `FEATURE_PREFIX=feature/`
- `HOTFIX_PREFIX=hotfix/`
- `AUTO_CREATE_MR=true`

## 与 gitlab-flow 的对比

| 特性 | gitlab-flow | flow3 |
|------|-------------|----------|
| 依赖 | 自维护 API 客户端 | 基于官方 glab |
| 配置复杂度 | 需要 OAuth 应用配置 | 使用 glab 认证 |
| 本地存储 | SQLite 数据库 | 无需本地存储 |
| 维护成本 | 需要跟进 GitLab API 变化 | 零维护 |
| 功能完整性 | 功能丰富但复杂 | 专注核心工作流 |

## 示例场景

### 场景1：开发新功能
```bash
# 1. 开始开发用户认证功能
flow3 feature start user-auth

# 2. 进行开发工作...
git add .
git commit -m "Add user authentication"

# 3. 完成开发
flow3 feature finish
# 自动推送分支并创建 MR
```

### 场景2：修复 Issue
```bash
# 1. 开始修复 Issue #456
flow-cli issue start 456

# 2. 修复代码...
git add .
git commit -m "Fix login validation bug"

# 3. 完成修复
flow-cli issue finish
# 自动创建 MR 并关联 Issue
```

## 故障排除

### glab 认证问题
```bash
# 检查 glab 认证状态
glab auth status

# 重新认证
glab auth login
```

### 分支冲突
```bash
# 手动解决冲突后继续
git add .
git commit -m "Resolve conflicts"
flow-cli feature finish
```

## 贡献

欢迎提交 Issue 和 PR 来改进 flow3。
