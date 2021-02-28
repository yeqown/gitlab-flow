## 自动补全

> 下面的文本主要来源于 https://github.com/urfave/cli/blob/master/docs/v2/manual.md#default-auto-completion


### 配置 Bash 自动补全

> 🤣 这部分我没有测试，所以遇到问题请先尝试自己解决（结合上述的链接）

1. 将下面的代码复制到一个文件中，如：`bash_auto_completion`。
2. 然后修改你的 `.bashrc` 文件：增加一行：`source path/to/bash_autocomplete`。
3. 在命令行中输入 `source ~/.bashrc` 使得刚刚的配置生效。

```sh
#! /bin/bash
PROG=gitlab-flow
: ${PROG:=$(basename ${BASH_SOURCE})}

_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if [[ "$cur" == "-"* ]]; then
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} ${cur} --generate-bash-completion )
    else
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    fi
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
  fi
}

complete -o bashdefault -o default -o nospace -F _cli_bash_autocomplete $PROG
unset PROG
```

### 配置 Zsh 自动补全

1. 将下面的代码复制到一个文件中，如：`zsh_auto_completion`.
2. 然后修改你的 `.zshrc` 文件：增加一行：`source path/to/zsh_autocomplete`.
3. 在命令行中输入 `source ~/.bashrc` 使得刚刚的配置生效。

```sh
# you can changes to your actual program name.
# As for me, I would like to use flow2 rather than `gitlab-flow`
PROG=gitlab-flow
_CLI_ZSH_AUTOCOMPLETE_HACK=1

#compdef $PROG

_cli_zsh_autocomplete() {
  local -a opts
  local cur
  cur=${words[-1]}
  if [[ "$cur" == "-"* ]]; then
    opts=("${(@f)$(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
  else
    opts=("${(@f)$(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:#words[@]-1} --generate-bash-completion)}")
  fi

  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi

  return
}

compdef _cli_zsh_autocomplete $PROG
```