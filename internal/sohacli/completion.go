package sohacli

import (
	"bytes"
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
		_, err := fmt.Fprint(rt.Out, buildBashCompletionScript())
		return err
	case "zsh":
		_, err := fmt.Fprint(rt.Out, "#compdef soha\n"+buildBashCompletionScript())
		return err
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

func buildBashCompletionScript() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "#!/usr/bin/env bash")
	fmt.Fprintln(&buf, "_soha_completion() {")
	fmt.Fprintln(&buf, "  local cur prev commands")
	fmt.Fprintln(&buf, "  COMPREPLY=()")
	fmt.Fprintln(&buf, "  cur=\"${COMP_WORDS[COMP_CWORD]}\"")
	fmt.Fprintln(&buf, "  prev=\"${COMP_WORDS[COMP_CWORD-1]}\"")
	fmt.Fprintf(&buf, "  commands=%q\n", joinWords(topLevelCommandNames(true)))
	fmt.Fprintln(&buf, "  case \"${COMP_WORDS[1]}\" in")
	for _, spec := range topLevelCommandSpecs {
		words := spec.completionWords()
		if len(words) == 0 {
			continue
		}
		fmt.Fprintf(&buf, "    %s)\n", spec.Name)
		writeCompletionSubcommands(&buf, spec)
		fmt.Fprintf(&buf, "      COMPREPLY=($(compgen -W %q -- \"$cur\"))\n", joinWords(words))
		fmt.Fprintln(&buf, "      return 0")
		fmt.Fprintln(&buf, "      ;;")
	}
	fmt.Fprintln(&buf, "  esac")
	fmt.Fprintln(&buf, "  COMPREPLY=($(compgen -W \"$commands\" -- \"$cur\"))")
	fmt.Fprintln(&buf, "}")
	fmt.Fprintln(&buf, "complete -F _soha_completion soha")
	return buf.String()
}

func writeCompletionSubcommands(buf *bytes.Buffer, spec commandSpec) {
	if len(spec.Subcommands) == 0 {
		return
	}
	for _, subcommand := range spec.Subcommands {
		words := subcommand.completionWords()
		if len(words) == 0 {
			continue
		}
		fmt.Fprintf(buf, "      if [[ \"${COMP_WORDS[2]}\" == %q ]]; then\n", subcommand.Name)
		fmt.Fprintf(buf, "        COMPREPLY=($(compgen -W %q -- \"$cur\"))\n", joinWords(words))
		fmt.Fprintln(buf, "        return 0")
		fmt.Fprintln(buf, "      fi")
	}
}
