package sohacli

import (
	"fmt"
	"io"
	"strings"
)

type commandDoc struct {
	Command  string
	Usage    string
	Purpose  string
	Examples []string
}

func runDocs(args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("docs", args, rt)
	format := fs.String("format", "markdown", "documentation output format: markdown")
	if err := fs.Parse(args); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "", "markdown":
		writeCommandDocsMarkdown(rt.Out)
		return nil
	default:
		return fmt.Errorf("unsupported docs format %q", *format)
	}
}

func writeCommandDocsMarkdown(out io.Writer) {
	fmt.Fprintln(out, "# Soha CLI Command Reference")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Generated with `soha docs --format markdown`.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "| Command | Usage | Purpose | Examples |")
	fmt.Fprintln(out, "| --- | --- | --- | --- |")
	for _, doc := range commandDocsFromSpecs() {
		fmt.Fprintf(
			out,
			"| `%s` | `%s` | %s | %s |\n",
			markdownCell(doc.Command),
			markdownCell(doc.Usage),
			markdownCell(doc.Purpose),
			markdownExamples(doc.Examples),
		)
	}
}

func commandDocsFromSpecs() []commandDoc {
	docs := make([]commandDoc, 0, len(topLevelCommandSpecs))
	for _, spec := range topLevelCommandSpecs {
		docs = appendCommandDocs(docs, "", spec)
	}
	return docs
}

func appendCommandDocs(docs []commandDoc, parent string, spec commandSpec) []commandDoc {
	command := strings.TrimSpace(strings.TrimSpace(parent + " " + spec.Name))
	if len(spec.Examples) > 0 {
		docs = append(docs, commandDoc{
			Command:  command,
			Usage:    spec.Usage,
			Purpose:  sentence(spec.Summary),
			Examples: append([]string(nil), spec.Examples...),
		})
	}
	for _, subcommand := range spec.Subcommands {
		docs = appendCommandDocs(docs, command, subcommand)
	}
	return docs
}

func sentence(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasSuffix(value, ".") {
		return value
	}
	return value + "."
}

func markdownExamples(examples []string) string {
	if len(examples) == 0 {
		return ""
	}
	out := make([]string, 0, len(examples))
	for _, example := range examples {
		out = append(out, "`"+markdownCell(example)+"`")
	}
	return strings.Join(out, "<br>")
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "|", `\|`)
	value = strings.ReplaceAll(value, "\n", "<br>")
	return value
}
