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

var commandDocs = []commandDoc{
	{
		Command:  "version",
		Usage:    "soha version [--json]",
		Purpose:  "Print build version information.",
		Examples: []string{"soha version", "soha version --json"},
	},
	{
		Command:  "login",
		Usage:    "soha login [options]",
		Purpose:  "Authenticate and store a local profile.",
		Examples: []string{"soha login --server http://localhost:8080 --login ada"},
	},
	{
		Command:  "capabilities",
		Usage:    "soha capabilities [--domain gateway|platform] [--output json|yaml|names|inputs]",
		Purpose:  "Print AI Gateway or platform capability metadata.",
		Examples: []string{"soha capabilities --output names", "soha capabilities --output inputs"},
	},
	{
		Command:  "tool call",
		Usage:    "soha tool call <name> [options]",
		Purpose:  "Invoke an AI Gateway tool with JSON input.",
		Examples: []string{"soha tool call k8s.pods.list --input-json '{\"clusterId\":\"local\"}'"},
	},
	{
		Command:  "resource read",
		Usage:    "soha resource read <uri> [options]",
		Purpose:  "Read an AI Gateway MCP resource.",
		Examples: []string{"soha resource read soha://k8s/runtime --context-json '{\"clusterId\":\"local\"}'"},
	},
	{
		Command:  "prompt get",
		Usage:    "soha prompt get <name> [options]",
		Purpose:  "Get an AI Gateway MCP prompt.",
		Examples: []string{"soha prompt get soha.k8s.diagnose_workload --arguments-json '{}'"},
	},
	{
		Command:  "token list",
		Usage:    "soha token list [options]",
		Purpose:  "List personal access tokens.",
		Examples: []string{"soha token list"},
	},
	{
		Command:  "token create",
		Usage:    "soha token create [options]",
		Purpose:  "Create a personal access token.",
		Examples: []string{"soha token create --name local --permission-keys ai.gateway.view"},
	},
	{
		Command:  "token revoke",
		Usage:    "soha token revoke <token-id> [options]",
		Purpose:  "Revoke a personal access token.",
		Examples: []string{"soha token revoke pat_123"},
	},
	{
		Command:  "service-account list",
		Usage:    "soha service-account list [options]",
		Purpose:  "List AI Gateway service accounts.",
		Examples: []string{"soha service-account list"},
	},
	{
		Command:  "service-account create",
		Usage:    "soha service-account create [options]",
		Purpose:  "Create a service account.",
		Examples: []string{"soha service-account create --name runner --role-ids operator"},
	},
	{
		Command:  "service-account token-list",
		Usage:    "soha service-account token-list [options]",
		Purpose:  "List service account tokens.",
		Examples: []string{"soha service-account token-list"},
	},
	{
		Command:  "service-account token-create",
		Usage:    "soha service-account token-create [options]",
		Purpose:  "Create a service account token.",
		Examples: []string{"soha service-account token-create --service-account-id sa_123 --name ci"},
	},
	{
		Command:  "service-account token-revoke",
		Usage:    "soha service-account token-revoke <token-id> [options]",
		Purpose:  "Revoke a service account token.",
		Examples: []string{"soha service-account token-revoke sat_123"},
	},
	{
		Command:  "audit list",
		Usage:    "soha audit list [options]",
		Purpose:  "Query AI Gateway audit logs.",
		Examples: []string{"soha audit list --tool-name k8s.pods.logs --limit 20"},
	},
	{
		Command:  "approval list",
		Usage:    "soha approval list [options]",
		Purpose:  "List approval requests.",
		Examples: []string{"soha approval list --status pending"},
	},
	{
		Command:  "approval timeline",
		Usage:    "soha approval timeline <approval-id> [options]",
		Purpose:  "Show an approval timeline.",
		Examples: []string{"soha approval timeline approval_123"},
	},
	{
		Command:  "approval approve",
		Usage:    "soha approval approve <approval-id> [options]",
		Purpose:  "Approve an approval request.",
		Examples: []string{"soha approval approve approval_123 --comment \"ship\""},
	},
	{
		Command:  "approval reject",
		Usage:    "soha approval reject <approval-id> [options]",
		Purpose:  "Reject an approval request.",
		Examples: []string{"soha approval reject approval_123 --comment \"needs review\""},
	},
	{
		Command:  "approval cancel",
		Usage:    "soha approval cancel <approval-id> [options]",
		Purpose:  "Cancel an approval request.",
		Examples: []string{"soha approval cancel approval_123"},
	},
	{
		Command:  "governance status",
		Usage:    "soha governance status [options]",
		Purpose:  "Show AI Gateway governance health and metrics.",
		Examples: []string{"soha governance status --window-hours 48", "soha governance status --json"},
	},
	{
		Command:  "cloud fleet diagnostics",
		Usage:    "soha cloud fleet diagnostics --tenant-id <tenant-id> --fleet-id <fleet-id> [--output summary|json|yaml]",
		Purpose:  "Show Cloud managed agent fleet capability diagnostics.",
		Examples: []string{"soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123", "soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123 --output json"},
	},
	{
		Command:  "profile list",
		Usage:    "soha profile list",
		Purpose:  "List local profiles.",
		Examples: []string{"soha profile list"},
	},
	{
		Command:  "profile show",
		Usage:    "soha profile show [profile]",
		Purpose:  "Show one local profile with tokens redacted.",
		Examples: []string{"soha profile show default"},
	},
	{
		Command:  "profile use",
		Usage:    "soha profile use <profile>",
		Purpose:  "Switch the current profile.",
		Examples: []string{"soha profile use cloud"},
	},
	{
		Command:  "context show",
		Usage:    "soha context show [options]",
		Purpose:  "Show AI client context headers.",
		Examples: []string{"soha context show"},
	},
	{
		Command:  "context set",
		Usage:    "soha context set [options]",
		Purpose:  "Update AI client context headers.",
		Examples: []string{"soha context set --skill-id k8s-sre --source codex"},
	},
	{
		Command:  "mcp start",
		Usage:    "soha mcp start [options]",
		Purpose:  "Run the Soha MCP stdio server.",
		Examples: []string{"soha mcp start --profile default"},
	},
	{
		Command:  "mcp install",
		Usage:    "soha mcp install [options]",
		Purpose:  "Print MCP client configuration.",
		Examples: []string{"soha mcp install --command /usr/local/bin/soha"},
	},
	{
		Command:  "skill list",
		Usage:    "soha skill list [options]",
		Purpose:  "List local Soha skill files.",
		Examples: []string{"soha skill list --source ../soha-skills"},
	},
	{
		Command:  "skill install",
		Usage:    "soha skill install [options] [skill-id...]",
		Purpose:  "Install local Soha skill files.",
		Examples: []string{"soha skill install --source ../soha-skills --dest ~/.soha/skills k8s-sre"},
	},
	{
		Command:  "add",
		Usage:    "soha add [target] [options]",
		Purpose:  "Add Soha MCP and skills to an AI agent or IDE.",
		Examples: []string{"soha add --profile local --server http://localhost:8080 --source ../soha-skills", "soha add claude --profile default --server https://mcp.opensoha.com --command /usr/local/bin/soha --source ../soha-skills"},
	},
	{
		Command:  "plugin search",
		Usage:    "soha plugin search [options]",
		Purpose:  "Search plugin marketplace entries.",
		Examples: []string{"soha plugin search --query k8s"},
	},
	{
		Command:  "plugin show",
		Usage:    "soha plugin show [plugin-id] [options]",
		Purpose:  "Show marketplace, installed, or manifest details.",
		Examples: []string{"soha plugin show opensoha.k8s-sre-pack --installed"},
	},
	{
		Command:  "plugin install",
		Usage:    "soha plugin install [plugin-id] [options]",
		Purpose:  "Install a marketplace or manifest plugin.",
		Examples: []string{"soha plugin install opensoha.k8s-sre-pack --enable"},
	},
	{
		Command:  "plugin list",
		Usage:    "soha plugin list [options]",
		Purpose:  "List installed plugins.",
		Examples: []string{"soha plugin list"},
	},
	{
		Command:  "plugin enable",
		Usage:    "soha plugin enable <plugin-id> [options]",
		Purpose:  "Enable an installed plugin.",
		Examples: []string{"soha plugin enable opensoha.k8s-sre-pack"},
	},
	{
		Command:  "plugin disable",
		Usage:    "soha plugin disable <plugin-id> [options]",
		Purpose:  "Disable an installed plugin.",
		Examples: []string{"soha plugin disable opensoha.k8s-sre-pack"},
	},
	{
		Command:  "plugin upgrade",
		Usage:    "soha plugin upgrade <plugin-id> [options]",
		Purpose:  "Upgrade an installed plugin.",
		Examples: []string{"soha plugin upgrade opensoha.k8s-sre-pack"},
	},
	{
		Command:  "plugin config",
		Usage:    "soha plugin config <plugin-id> [options]",
		Purpose:  "Configure metadata and secret refs.",
		Examples: []string{"soha plugin config opensoha.k8s-sre-pack --secret-ref token=secret://k8s/token"},
	},
	{
		Command:  "plugin remove",
		Usage:    "soha plugin remove <plugin-id> [options]",
		Purpose:  "Remove an installed plugin.",
		Examples: []string{"soha plugin remove opensoha.k8s-sre-pack"},
	},
	{
		Command:  "diagnose",
		Usage:    "soha diagnose [options]",
		Purpose:  "Check profile and Gateway visibility.",
		Examples: []string{"soha diagnose --tool k8s.pods.logs --resource soha://k8s/runtime"},
	},
	{
		Command:  "completion",
		Usage:    "soha completion [bash|zsh]",
		Purpose:  "Print shell completion script.",
		Examples: []string{"soha completion bash", "soha completion zsh"},
	},
	{
		Command:  "docs",
		Usage:    "soha docs [--format markdown]",
		Purpose:  "Generate CLI command reference documentation.",
		Examples: []string{"soha docs --format markdown"},
	},
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
	for _, doc := range commandDocs {
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
