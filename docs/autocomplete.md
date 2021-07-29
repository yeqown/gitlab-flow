## Autocomplete

> All paragraph of following is from https://github.com/urfave/cli/blob/master/docs/v2/manual.md#default-auto-completion


### Bash auto-completion

> 🤣 I did not test this, any issue you got may need to resolve it by yourself.

1. Copy following shell into a file `zsh_auto_completion`.
2. Then modify your `.bashrc` to add `source path/to/zsh_autocomplete`.
3. `source ~/.zshrc` to take it effect.

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

### Zsh auto-completion

1. Copy following shell into a file `zsh_auto_completion`.
2. Then modify your `.zshrc` to add `source path/to/zsh_autocomplete`.
3. `source ~/.zshrc` to take it effect.

```sh
# you can changes to your actual program name.
# As for me, I would like to use flows rather than `gitlab-flow`
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