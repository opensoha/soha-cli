package sohacli

import (
	"context"
	"fmt"
	"strings"
)

type commandHandler func(context.Context, []string, Runtime) error

type commandSpec struct {
	Name            string
	Aliases         []string
	Usage           string
	Summary         string
	Examples        []string
	CompletionWords []string
	Subcommands     []commandSpec
	Handler         commandHandler
}

var topLevelCommandSpecs = []commandSpec{
	{
		Name:     "version",
		Usage:    "soha version [--json]",
		Summary:  "Print build version information",
		Examples: []string{"soha version", "soha version --json"},
		Handler:  func(_ context.Context, args []string, rt Runtime) error { return runVersion(args, rt) },
	},
	{
		Name:     "login",
		Usage:    "soha login [options]",
		Summary:  "Authenticate and store a local profile",
		Examples: []string{"soha login --server http://localhost:8080 --login ada"},
		Handler:  runLogin,
	},
	{
		Name:     "capabilities",
		Usage:    "soha capabilities [--domain gateway|platform] [--output json|yaml|names|inputs]",
		Summary:  "Print AI Gateway or platform capability metadata",
		Examples: []string{"soha capabilities --output names", "soha capabilities --output inputs"},
		Handler:  runCapabilities,
	},
	{
		Name:    "tool",
		Usage:   "soha tool call <name> [options]",
		Summary: "Invoke an AI Gateway tool with JSON input",
		Subcommands: []commandSpec{
			{
				Name:     "call",
				Usage:    "soha tool call <name> [options]",
				Summary:  "Invoke an AI Gateway tool with JSON input",
				Examples: []string{"soha tool call k8s.pods.list --input-json '{\"clusterId\":\"local\"}'"},
			},
		},
		Handler: runTool,
	},
	{
		Name:    "resource",
		Usage:   "soha resource read <uri> [options]",
		Summary: "Read an AI Gateway MCP resource",
		Subcommands: []commandSpec{
			{
				Name:     "read",
				Usage:    "soha resource read <uri> [options]",
				Summary:  "Read an AI Gateway MCP resource",
				Examples: []string{"soha resource read soha://k8s/runtime --context-json '{\"clusterId\":\"local\"}'"},
			},
		},
		Handler: runResource,
	},
	{
		Name:    "prompt",
		Usage:   "soha prompt get <name> [options]",
		Summary: "Get an AI Gateway MCP prompt",
		Subcommands: []commandSpec{
			{
				Name:     "get",
				Usage:    "soha prompt get <name> [options]",
				Summary:  "Get an AI Gateway MCP prompt",
				Examples: []string{"soha prompt get soha.k8s.diagnose_workload --arguments-json '{}'"},
			},
		},
		Handler: runPrompt,
	},
	{
		Name:    "token",
		Usage:   "soha token <list|create|revoke> [options]",
		Summary: "Manage personal access tokens",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha token list [options]", Summary: "List personal access tokens", Examples: []string{"soha token list"}},
			{Name: "create", Usage: "soha token create [options]", Summary: "Create a personal access token", Examples: []string{"soha token create --name local --permission-keys ai.gateway.view"}},
			{Name: "revoke", Usage: "soha token revoke <token-id> [options]", Summary: "Revoke a personal access token", Examples: []string{"soha token revoke pat_123"}},
		},
		Handler: runToken,
	},
	{
		Name:    "service-account",
		Usage:   "soha service-account <list|create|token-list|token-create|token-revoke> [options]",
		Summary: "Manage AI Gateway service accounts and tokens",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha service-account list [options]", Summary: "List AI Gateway service accounts", Examples: []string{"soha service-account list"}},
			{Name: "create", Usage: "soha service-account create [options]", Summary: "Create a service account", Examples: []string{"soha service-account create --name runner --role-ids operator"}},
			{Name: "token-list", Usage: "soha service-account token-list [options]", Summary: "List service account tokens", Examples: []string{"soha service-account token-list"}},
			{Name: "token-create", Usage: "soha service-account token-create [options]", Summary: "Create a service account token", Examples: []string{"soha service-account token-create --service-account-id sa_123 --name ci"}},
			{Name: "token-revoke", Usage: "soha service-account token-revoke <token-id> [options]", Summary: "Revoke a service account token", Examples: []string{"soha service-account token-revoke sat_123"}},
		},
		Handler: runServiceAccount,
	},
	{
		Name:    "audit",
		Usage:   "soha audit list [options]",
		Summary: "Query AI Gateway audit logs",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha audit list [options]", Summary: "Query AI Gateway audit logs", Examples: []string{"soha audit list --tool-name k8s.pods.logs --limit 20"}},
		},
		Handler: runAudit,
	},
	{
		Name:    "approval",
		Usage:   "soha approval <list|timeline|approve|reject|cancel> [options]",
		Summary: "List, trace, or decide AI Gateway approval requests",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha approval list [options]", Summary: "List approval requests", Examples: []string{"soha approval list --status pending"}},
			{Name: "timeline", Usage: "soha approval timeline <approval-id> [options]", Summary: "Show an approval timeline", Examples: []string{"soha approval timeline approval_123"}},
			{Name: "approve", Usage: "soha approval approve <approval-id> [options]", Summary: "Approve an approval request", Examples: []string{"soha approval approve approval_123 --comment \"ship\""}},
			{Name: "reject", Usage: "soha approval reject <approval-id> [options]", Summary: "Reject an approval request", Examples: []string{"soha approval reject approval_123 --comment \"needs review\""}},
			{Name: "cancel", Usage: "soha approval cancel <approval-id> [options]", Summary: "Cancel an approval request", Examples: []string{"soha approval cancel approval_123"}},
		},
		Handler: runApproval,
	},
	{
		Name:    "governance",
		Usage:   "soha governance status [options]",
		Summary: "Show AI Gateway governance health and metrics",
		Subcommands: []commandSpec{
			{Name: "status", Usage: "soha governance status [options]", Summary: "Show AI Gateway governance health and metrics", Examples: []string{"soha governance status --window-hours 48", "soha governance status --json"}},
		},
		Handler: runGovernance,
	},
	{
		Name:    "cloud",
		Usage:   "soha cloud fleet diagnostics [options]",
		Summary: "Show Cloud managed agent fleet capability diagnostics",
		Subcommands: []commandSpec{
			{
				Name:    "fleet",
				Usage:   "soha cloud fleet diagnostics [options]",
				Summary: "Manage Cloud agent fleet diagnostics",
				Subcommands: []commandSpec{
					{
						Name:     "diagnostics",
						Usage:    "soha cloud fleet diagnostics --tenant-id <tenant-id> --fleet-id <fleet-id> [--output summary|json|yaml]",
						Summary:  "Show Cloud managed agent fleet capability diagnostics",
						Examples: []string{"soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123", "soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123 --output json"},
					},
				},
			},
		},
		Handler: runCloud,
	},
	{
		Name:    "profile",
		Usage:   "soha profile <list|show|use> [options]",
		Summary: "List, show, or switch profiles",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha profile list", Summary: "List local profiles", Examples: []string{"soha profile list"}},
			{Name: "show", Usage: "soha profile show [profile]", Summary: "Show one local profile with tokens redacted", Examples: []string{"soha profile show default"}},
			{Name: "use", Usage: "soha profile use <profile>", Summary: "Switch the current profile", Examples: []string{"soha profile use cloud"}},
		},
		Handler: func(_ context.Context, args []string, rt Runtime) error { return runProfile(args, rt) },
	},
	{
		Name:    "context",
		Usage:   "soha context <show|set> [options]",
		Summary: "Show or update AI client context headers and marketplace defaults",
		Subcommands: []commandSpec{
			{Name: "show", Usage: "soha context show [options]", Summary: "Show AI client context headers and marketplace defaults", Examples: []string{"soha context show"}},
			{Name: "set", Usage: "soha context set [options]", Summary: "Update AI client context headers and marketplace defaults", Examples: []string{"soha context set --skill-id k8s-sre --source codex", "soha context set --marketplace https://marketplace.opensoha.com --marketplace-source-id opensoha-official"}},
		},
		Handler: runContext,
	},
	{
		Name:    "mcp",
		Usage:   "soha mcp <start|install> [options]",
		Summary: "Run or install the soha MCP stdio server",
		Subcommands: []commandSpec{
			{Name: "start", Usage: "soha mcp start [options]", Summary: "Run the Soha MCP stdio server", Examples: []string{"soha mcp start --profile default"}},
			{Name: "install", Usage: "soha mcp install [options]", Summary: "Print MCP client configuration", Examples: []string{"soha mcp install --command /usr/local/bin/soha"}},
		},
		Handler: runMCP,
	},
	{
		Name:    "skill",
		Usage:   "soha skill <list|install> [options]",
		Summary: "List or install local soha AI Gateway skill files",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha skill list [options]", Summary: "List local Soha skill files", Examples: []string{"soha skill list --source ../soha-skills"}},
			{Name: "install", Usage: "soha skill install [options] [skill-id...]", Summary: "Install local Soha skill files", Examples: []string{"soha skill install --source ../soha-skills --dest ~/.soha/skills k8s-sre"}},
		},
		Handler: func(_ context.Context, args []string, rt Runtime) error { return runSkill(args, rt) },
	},
	{
		Name:            "add",
		Usage:           "soha add [target] [options]",
		Summary:         "Add Soha MCP and skills to an AI agent or IDE",
		Examples:        []string{"soha add --profile local --server http://localhost:8080 --source ../soha-skills", "soha add claude --profile default --server https://mcp.opensoha.com --command /usr/local/bin/soha --source ../soha-skills"},
		CompletionWords: []string{"codex", "claude", "cursor", "kiro", "gemini", "antigravity", "antigravity-ide", "trae", "all"},
		Handler:         func(_ context.Context, args []string, rt Runtime) error { return runAdd(args, rt) },
	},
	{
		Name:    "plugin",
		Usage:   "soha plugin <search|show|install|list|enable|disable|upgrade|config|remove> [options]",
		Summary: "Search, install, and manage Soha plugins",
		Subcommands: []commandSpec{
			{Name: "search", Usage: "soha plugin search [options]", Summary: "Search plugin marketplace entries", Examples: []string{"soha plugin search --query k8s", "soha plugin search --marketplace https://marketplace.opensoha.com --source-id opensoha-official"}},
			{Name: "show", Usage: "soha plugin show [options] <plugin-id>", Summary: "Show marketplace, installed, or manifest details", Examples: []string{"soha plugin show --installed opensoha.k8s-sre-pack", "soha plugin show --marketplace https://marketplace.opensoha.com --version 0.1.0 opensoha.k8s-sre-pack"}},
			{Name: "install", Usage: "soha plugin install [options] [plugin-id]", Summary: "Install a marketplace or manifest plugin", Examples: []string{"soha plugin install --enable opensoha.k8s-sre-pack", "soha plugin install --marketplace https://marketplace.opensoha.com --version 0.1.0 opensoha.k8s-sre-pack"}},
			{Name: "list", Usage: "soha plugin list [options]", Summary: "List installed plugins", Examples: []string{"soha plugin list"}},
			{Name: "enable", Usage: "soha plugin enable <plugin-id> [options]", Summary: "Enable an installed plugin", Examples: []string{"soha plugin enable opensoha.k8s-sre-pack"}},
			{Name: "disable", Usage: "soha plugin disable <plugin-id> [options]", Summary: "Disable an installed plugin", Examples: []string{"soha plugin disable opensoha.k8s-sre-pack"}},
			{Name: "upgrade", Usage: "soha plugin upgrade [options] <plugin-id>", Summary: "Upgrade an installed plugin", Examples: []string{"soha plugin upgrade opensoha.k8s-sre-pack", "soha plugin upgrade --marketplace https://marketplace.opensoha.com --version 0.2.0 opensoha.k8s-sre-pack"}},
			{Name: "config", Usage: "soha plugin config <plugin-id> [options]", Summary: "Configure metadata and secret refs", Examples: []string{"soha plugin config opensoha.k8s-sre-pack --secret-ref token=secret://k8s/token"}},
			{Name: "remove", Aliases: []string{"rm"}, Usage: "soha plugin remove <plugin-id> [options]", Summary: "Remove an installed plugin", Examples: []string{"soha plugin remove opensoha.k8s-sre-pack"}},
		},
		Handler: runPlugin,
	},
	{
		Name:     "diagnose",
		Usage:    "soha diagnose [options]",
		Summary:  "Check profile and Gateway connectivity",
		Examples: []string{"soha diagnose --tool k8s.pods.logs --resource soha://k8s/runtime"},
		Handler:  runDiagnose,
	},
	{
		Name:            "completion",
		Usage:           "soha completion [bash|zsh]",
		Summary:         "Print shell completion script",
		Examples:        []string{"soha completion bash", "soha completion zsh"},
		CompletionWords: []string{"bash", "zsh"},
	},
	{
		Name:     "docs",
		Usage:    "soha docs [--format markdown]",
		Summary:  "Generate CLI command reference documentation",
		Examples: []string{"soha docs --format markdown"},
	},
}

func findTopLevelCommandSpec(name string) (*commandSpec, bool) {
	for i := range topLevelCommandSpecs {
		if topLevelCommandSpecs[i].matches(name) {
			return &topLevelCommandSpecs[i], true
		}
	}
	return nil, false
}

func dispatchTopLevelCommand(ctx context.Context, name string, args []string, rt Runtime) error {
	spec, ok := findTopLevelCommandSpec(name)
	if !ok {
		return fmt.Errorf("unknown command %q", name)
	}
	if spec.Handler == nil {
		switch spec.Name {
		case "completion":
			return runCompletion(args, rt)
		case "docs":
			return runDocs(args, rt)
		default:
			return fmt.Errorf("unknown command %q", name)
		}
	}
	return spec.Handler(ctx, args, rt)
}

func topLevelCommandNames(includeHelp bool) []string {
	names := make([]string, 0, len(topLevelCommandSpecs)+1)
	for _, spec := range topLevelCommandSpecs {
		names = append(names, spec.Name)
		names = append(names, spec.Aliases...)
	}
	if includeHelp {
		names = append(names, "help")
	}
	return names
}

func (s commandSpec) matches(name string) bool {
	if s.Name == name {
		return true
	}
	for _, alias := range s.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func (s commandSpec) completionWords() []string {
	if len(s.CompletionWords) > 0 {
		return append([]string(nil), s.CompletionWords...)
	}
	words := make([]string, 0, len(s.Subcommands))
	for _, subcommand := range s.Subcommands {
		words = append(words, subcommand.Name)
		words = append(words, subcommand.Aliases...)
	}
	return words
}

func joinWords(words []string) string {
	return strings.Join(words, " ")
}
