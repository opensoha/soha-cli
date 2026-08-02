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
		Name:    "logs",
		Usage:   "soha logs <query|tail> [options]",
		Summary: "Query or follow operational logs",
		Subcommands: []commandSpec{
			{Name: "query", Usage: "soha logs query --source <cluster|docker|delivery> [options]", Summary: "Query a bounded page of logs", Examples: []string{"soha logs query --source cluster --cluster-id local --namespace default --output json", "soha logs query --source docker --project-id project-1 --service api --output ndjson"}},
			{Name: "tail", Usage: "soha logs tail --source <cluster|docker|delivery> [options]", Summary: "Follow logs as NDJSON", Examples: []string{"soha logs tail --source cluster --cluster-id local --namespace default", "soha logs tail --source delivery --application-id app-1 --environment-id env-1"}},
		},
		Handler: runLogs,
	},
	{
		Name:    "operation",
		Usage:   "soha operation <get|wait|cancel> <domain> <id> [options]",
		Summary: "Inspect and control asynchronous operations",
		Subcommands: []commandSpec{
			{Name: "get", Usage: "soha operation get <virtualization|container_runtime> <id> [options]", Summary: "Get an operation", Examples: []string{"soha operation get virtualization task-1 --output json"}},
			{Name: "wait", Usage: "soha operation wait <virtualization|container_runtime> <id> [options]", Summary: "Wait for an operation to finish", Examples: []string{"soha operation wait container_runtime task-1 --wait-timeout 10m --output json"}},
			{Name: "cancel", Usage: "soha operation cancel <virtualization|container_runtime> <id> [options]", Summary: "Cancel an operation", Examples: []string{"soha operation cancel virtualization task-1 --yes"}},
		},
		Handler: runOperation,
	},
	{
		Name:    "tool",
		Usage:   "soha tool call <name> [--preview] [--yes] [options]",
		Summary: "Invoke an AI Gateway tool with JSON input",
		Subcommands: []commandSpec{
			{
				Name:     "call",
				Usage:    "soha tool call <name> [--preview] [--yes] [options]",
				Summary:  "Invoke an AI Gateway tool with JSON input",
				Examples: []string{"soha tool call k8s.pods.list --input-json '{\"clusterId\":\"local\"}'", "soha tool call k8s.deployments.restart --preview --input-json '{\"clusterId\":\"local\"}'", "soha tool call k8s.deployments.restart --yes --input-json '{\"clusterId\":\"local\"}'"},
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
		Name:    "knowledge",
		Usage:   "soha knowledge <search|connectors|sync|rebuild> [options]",
		Summary: "Search and operate authorized AI knowledge pipelines",
		Subcommands: []commandSpec{
			{
				Name:     "search",
				Usage:    "soha knowledge search --base-ids <id[,id...]> --query <text> [options]",
				Summary:  "Search authorized knowledge bases and return grounded evidence",
				Examples: []string{"soha knowledge search --base-ids runbooks --query \"deployment rollback\"", "soha knowledge search --base-ids runbooks,handbook --query \"incident owner\" --output json"},
			},
			{Name: "connectors", Usage: "soha knowledge connectors <list|create|validate> [options]", Summary: "List, create, or validate external knowledge connectors", Examples: []string{"soha knowledge connectors list", "soha knowledge connectors create --base-id runbooks --name docs --kind http --config-ref secret:knowledge/docs --config-json '{\"url\":\"https://docs.example.com/\",\"allowedHosts\":[\"docs.example.com\"],\"maxBytes\":8388608}'", "soha knowledge connectors validate connector-1"}, CompletionWords: []string{"list", "create", "validate"}},
			{Name: "sync", Usage: "soha knowledge sync <start|status|cancel|retry> [options]", Summary: "Start or control bounded knowledge synchronization jobs", Examples: []string{"soha knowledge sync start --base-id runbooks --source-id source-1", "soha knowledge sync status sync-1"}, CompletionWords: []string{"start", "status", "cancel", "retry"}},
			{Name: "rebuild", Usage: "soha knowledge rebuild --base-id <id> [options]", Summary: "Start a bounded knowledge index rebuild", Examples: []string{"soha knowledge rebuild --base-id runbooks --reason \"embedding model update\""}},
		},
		Handler: runKnowledge,
	},
	{
		Name:    "ai",
		Usage:   "soha ai <evaluation|memory> [options]",
		Summary: "Operate AI evaluation and governed memory workflows",
		Subcommands: []commandSpec{
			{Name: "evaluation", Usage: "soha ai evaluation <run|replay|gate> [options]", Summary: "Execute evaluation runs, isolated replay, or release gates", Examples: []string{"soha ai evaluation run --executor-profile-id executor-1 eval-run-1", "soha ai evaluation replay --id replay-1 --baseline-run-id baseline-1 --candidate-run-id candidate-1 --executor-profile-id executor-1", "soha ai evaluation gate --policy-id policy-1 --baseline-run-id baseline-1 --candidate-run-id candidate-1"}, CompletionWords: []string{"run", "replay", "gate"}},
			{Name: "memory", Usage: "soha ai memory <inspect|delete> [options]", Summary: "Inspect or delete governed memory records", Examples: []string{"soha ai memory inspect --owner-type user --owner-id user-1", "soha ai memory delete memory-1"}, CompletionWords: []string{"inspect", "delete"}},
		},
		Handler: runAI,
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
		Usage:   "soha mcp [options] | soha mcp <start|install> [options]",
		Summary: "Run the Soha MCP stdio server or print client configuration",
		Subcommands: []commandSpec{
			{Name: "start", Usage: "soha mcp start [options]", Summary: "Compatibility form of soha mcp", Examples: []string{"soha mcp start --profile default"}},
			{Name: "install", Usage: "soha mcp install [options]", Summary: "Print MCP client configuration", Examples: []string{"soha mcp install --command /usr/local/bin/soha", "soha mcp install --base-url https://soha.example.com"}},
		},
		Examples: []string{"soha mcp", "soha mcp --base-url https://soha.example.com"},
		Handler:  runMCP,
	},
	{
		Name:    "skill",
		Usage:   "soha skill <list|install|status|update|remove|rollback> [options]",
		Summary: "Install and manage verified Soha AI Gateway skills",
		Subcommands: []commandSpec{
			{Name: "list", Usage: "soha skill list [options]", Summary: "List skills from a local or verified release source", Examples: []string{"soha skill list"}},
			{Name: "install", Usage: "soha skill install [options] [skill-id...]", Summary: "Install skills from a local or verified release source", Examples: []string{"soha skill install --dest ~/.soha/skills k8s-sre"}},
			{Name: "status", Usage: "soha skill status [options]", Summary: "Show the active and previous managed skills generations", Examples: []string{"soha skill status", "soha skill status --scope project --json"}},
			{Name: "update", Usage: "soha skill update [options] [skill-id...]", Summary: "Update managed skills from the latest stable release", Examples: []string{"soha skill update", "soha skill update --all"}},
			{Name: "remove", Usage: "soha skill remove [options] <skill-id...>", Summary: "Remove managed skills with a rollback generation", Examples: []string{"soha skill remove k8s-sre"}},
			{Name: "rollback", Usage: "soha skill rollback [options]", Summary: "Switch back to the previous verified skills generation", Examples: []string{"soha skill rollback"}},
		},
		Handler: runSkill,
	},
	{
		Name:            "setup",
		Usage:           "soha setup [target] [--client target] [--mode mcp|skill|both] [--scope user|project] [options]",
		Summary:         "Configure Soha MCP and skills for an AI agent or IDE",
		Examples:        []string{"soha setup --client codex", "soha setup --client codex --scope project", "soha setup --client claude --mode mcp --base-url https://soha.example.com", "soha setup --client codex --check"},
		CompletionWords: []string{"codex", "claude", "cursor", "kiro", "gemini", "antigravity", "antigravity-ide", "trae", "all"},
		Handler:         runSetup,
	},
	{
		Name:            "add",
		Usage:           "soha add [target] [options]",
		Summary:         "Compatibility alias for configuring Soha MCP and skills",
		Examples:        []string{"soha add --profile local --server http://localhost:8080", "soha add claude --profile default --server https://mcp.opensoha.com --command /usr/local/bin/soha"},
		CompletionWords: []string{"codex", "claude", "cursor", "kiro", "gemini", "antigravity", "antigravity-ide", "trae", "all"},
		Handler:         runAdd,
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
		Usage:    "soha diagnose [--client <name>] [--output text|json] [options]",
		Summary:  "Check profile, client setup, and Gateway connectivity",
		Examples: []string{"soha diagnose --client codex --output json", "soha diagnose --tool k8s.pods.logs --resource soha://k8s/runtime"},
		Handler:  runDiagnose,
	},
	{
		Name:            "completion",
		Usage:           "soha completion [bash|zsh|fish|powershell]",
		Summary:         "Print shell completion script",
		Examples:        []string{"soha completion bash", "soha completion zsh", "soha completion fish", "soha completion powershell"},
		CompletionWords: []string{"bash", "zsh", "fish", "powershell"},
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

func findCommandSpec(path []string) (*commandSpec, bool) {
	if len(path) == 0 {
		return nil, false
	}
	spec, ok := findTopLevelCommandSpec(path[0])
	if !ok {
		return nil, false
	}
	for _, name := range path[1:] {
		if isHelpArg(name) || strings.HasPrefix(name, "-") {
			break
		}
		var next *commandSpec
		for i := range spec.Subcommands {
			if spec.Subcommands[i].matches(name) {
				next = &spec.Subcommands[i]
				break
			}
		}
		if next == nil {
			return nil, false
		}
		spec = next
	}
	return spec, true
}

func dispatchTopLevelCommand(ctx context.Context, name string, args []string, rt Runtime) error {
	spec, ok := findTopLevelCommandSpec(name)
	if !ok {
		return usageError{message: fmt.Sprintf("unknown command %q", name)}
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
