package sohacli

import (
	"fmt"
	"strings"
)

func runCompletion(args []string, rt Runtime) error {
	shell := "bash"
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		shell = strings.TrimSpace(args[0])
	}
	switch shell {
	case "bash":
		_, err := fmt.Fprint(rt.Out, bashCompletionScript)
		return err
	case "zsh":
		_, err := fmt.Fprint(rt.Out, "#compdef soha\n"+bashCompletionScript)
		return err
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

const bashCompletionScript = `#!/usr/bin/env bash
_soha_completion() {
  local cur prev commands
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="version login capabilities tool resource prompt token service-account audit approval governance cloud profile context mcp skill add plugin diagnose completion docs help"
  case "${COMP_WORDS[1]}" in
    tool)
      COMPREPLY=($(compgen -W "call" -- "$cur"))
      return 0
      ;;
    resource)
      COMPREPLY=($(compgen -W "read" -- "$cur"))
      return 0
      ;;
    prompt)
      COMPREPLY=($(compgen -W "get" -- "$cur"))
      return 0
      ;;
    token)
      COMPREPLY=($(compgen -W "list create revoke" -- "$cur"))
      return 0
      ;;
    service-account)
      COMPREPLY=($(compgen -W "list create token-list token-create token-revoke" -- "$cur"))
      return 0
      ;;
    audit)
      COMPREPLY=($(compgen -W "list" -- "$cur"))
      return 0
      ;;
    approval)
      COMPREPLY=($(compgen -W "list timeline approve reject cancel" -- "$cur"))
      return 0
      ;;
    governance)
      COMPREPLY=($(compgen -W "status" -- "$cur"))
      return 0
      ;;
    cloud)
      if [[ "${COMP_WORDS[2]}" == "fleet" ]]; then
        COMPREPLY=($(compgen -W "diagnostics" -- "$cur"))
        return 0
      fi
      COMPREPLY=($(compgen -W "fleet" -- "$cur"))
      return 0
      ;;
    profile)
      COMPREPLY=($(compgen -W "list show use" -- "$cur"))
      return 0
      ;;
    context)
      COMPREPLY=($(compgen -W "show set" -- "$cur"))
      return 0
      ;;
    mcp)
      COMPREPLY=($(compgen -W "start install" -- "$cur"))
      return 0
      ;;
    skill)
      COMPREPLY=($(compgen -W "list install" -- "$cur"))
      return 0
      ;;
    add)
      COMPREPLY=($(compgen -W "codex claude cursor kiro gemini antigravity antigravity-ide trae all" -- "$cur"))
      return 0
      ;;
    plugin)
      COMPREPLY=($(compgen -W "search show install list enable disable upgrade config remove rm" -- "$cur"))
      return 0
      ;;
  esac
  COMPREPLY=($(compgen -W "$commands" -- "$cur"))
}
complete -F _soha_completion soha
`
