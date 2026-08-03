# Soha CLI Command Reference

Generated with `soha docs --format markdown`.

## Automation contract

Structured results are written to stdout. Diagnostics and prompts are written to stderr. Exit codes are stable: `0` success/help, `1` validation, execution, API, or terminal operation failure, `2` missing/unknown top-level command or malformed flag, and `130` interrupted execution.

| Command | Usage | Purpose | Examples |
| --- | --- | --- | --- |
| `version` | `soha version [--json]` | Print build version information. | `soha version`<br>`soha version --json` |
| `login` | `soha login [options]` | Authenticate and store a local profile. | `soha login --server http://localhost:8080 --login ada` |
| `capabilities` | `soha capabilities [--domain gateway\|platform] [--output json\|yaml\|names\|inputs]` | Print AI Gateway or platform capability metadata. | `soha capabilities --output names`<br>`soha capabilities --output inputs` |
| `logs query` | `soha logs query --source <cluster\|docker\|delivery> [options]` | Query a bounded page of logs. | `soha logs query --source cluster --cluster-id local --namespace default --output json`<br>`soha logs query --source docker --project-id project-1 --service api --output ndjson` |
| `logs tail` | `soha logs tail --source <cluster\|docker\|delivery> [options]` | Follow logs as NDJSON. | `soha logs tail --source cluster --cluster-id local --namespace default`<br>`soha logs tail --source delivery --application-id app-1 --environment-id env-1` |
| `operation get` | `soha operation get <virtualization\|container_runtime> <id> [options]` | Get an operation. | `soha operation get virtualization task-1 --output json` |
| `operation wait` | `soha operation wait <virtualization\|container_runtime> <id> [options]` | Wait for an operation to finish. | `soha operation wait container_runtime task-1 --wait-timeout 10m --output json` |
| `operation cancel` | `soha operation cancel <virtualization\|container_runtime> <id> [options]` | Cancel an operation. | `soha operation cancel virtualization task-1 --yes` |
| `tool call` | `soha tool call <name> [--preview] [--yes] [options]` | Invoke an AI Gateway tool with JSON input. | `soha tool call k8s.pods.list --input-json '{"clusterId":"local"}'`<br>`soha tool call k8s.deployments.restart --preview --input-json '{"clusterId":"local"}'`<br>`soha tool call k8s.deployments.restart --yes --input-json '{"clusterId":"local"}'` |
| `project plan` | `soha project plan [file] [--output json\|yaml] [options]` | Validate a project manifest and invoke side-effect-free plans. | `soha project plan`<br>`soha project plan .soha/project.yaml --output yaml` |
| `project apply` | `soha project apply [file] [--yes] [options]` | Apply project steps in dependency order. | `soha project apply .soha/project.yaml`<br>`soha project apply --yes` |
| `resource read` | `soha resource read <uri> [options]` | Read an AI Gateway MCP resource. | `soha resource read soha://k8s/runtime --context-json '{"clusterId":"local"}'` |
| `knowledge search` | `soha knowledge search --base-ids <id[,id...]> --query <text> [options]` | Search authorized knowledge bases and return grounded evidence. | `soha knowledge search --base-ids runbooks --query "deployment rollback"`<br>`soha knowledge search --base-ids runbooks,handbook --query "incident owner" --output json` |
| `knowledge connectors` | `soha knowledge connectors <list\|create\|validate> [options]` | List, create, or validate external knowledge connectors. | `soha knowledge connectors list`<br>`soha knowledge connectors create --base-id runbooks --name docs --kind http --config-ref secret:knowledge/docs --config-json '{"url":"https://docs.example.com/","allowedHosts":["docs.example.com"],"maxBytes":8388608}'`<br>`soha knowledge connectors validate connector-1` |
| `knowledge sync` | `soha knowledge sync <start\|status\|cancel\|retry> [options]` | Start or control bounded knowledge synchronization jobs. | `soha knowledge sync start --base-id runbooks --source-id source-1`<br>`soha knowledge sync status sync-1` |
| `knowledge rebuild` | `soha knowledge rebuild --base-id <id> [options]` | Start a bounded knowledge index rebuild. | `soha knowledge rebuild --base-id runbooks --reason "embedding model update"` |
| `ai evaluation` | `soha ai evaluation <run\|replay\|gate> [options]` | Execute evaluation runs, isolated replay, or release gates. | `soha ai evaluation run --executor-profile-id executor-1 eval-run-1`<br>`soha ai evaluation replay --id replay-1 --baseline-run-id baseline-1 --candidate-run-id candidate-1 --executor-profile-id executor-1`<br>`soha ai evaluation gate --policy-id policy-1 --baseline-run-id baseline-1 --candidate-run-id candidate-1` |
| `ai memory` | `soha ai memory <inspect\|delete> [options]` | Inspect or delete governed memory records. | `soha ai memory inspect --owner-type user --owner-id user-1`<br>`soha ai memory delete memory-1` |
| `prompt get` | `soha prompt get <name> [options]` | Get an AI Gateway MCP prompt. | `soha prompt get soha.k8s.diagnose_workload --arguments-json '{}'` |
| `secret list` | `soha secret list [--scope-type type --scope-id id] [options]` | List authorized secret metadata. | `soha secret list --scope-type project --scope-id demo` |
| `secret get` | `soha secret get <secret-id> [options]` | Get secret metadata without revealing its value. | `soha secret get registry-token` |
| `secret create` | `soha secret create --name <name> [--binding type=target] [options]` | Create a secret by reading its value without terminal echo. | `printf '%s\n' "$REGISTRY_TOKEN" \| soha secret create --name registry-token --scope-type project --scope-id demo --binding capability=docker.projects.deploy.trigger` |
| `secret update` | `soha secret update <secret-id> [options]` | Update secret metadata and bindings. | `soha secret update registry-token --binding capability=docker.projects.deploy.trigger` |
| `secret disable` | `soha secret disable <secret-id> [options]` | Disable a secret. | `soha secret disable registry-token` |
| `secret versions` | `soha secret versions <secret-id> [options]` | List immutable secret versions. | `soha secret versions registry-token` |
| `secret rotate` | `soha secret rotate <secret-id> [options]` | Create a new secret version from a hidden value. | `printf '%s\n' "$REGISTRY_TOKEN" \| soha secret rotate registry-token` |
| `secret revoke-version` | `soha secret revoke-version <secret-id> <version> [options]` | Revoke a secret version. | `soha secret revoke-version registry-token 1` |
| `token list` | `soha token list [options]` | List personal access tokens. | `soha token list` |
| `token create` | `soha token create [options]` | Create a personal access token. | `soha token create --name local --permission-keys ai.gateway.view` |
| `token revoke` | `soha token revoke <token-id> [options]` | Revoke a personal access token. | `soha token revoke pat_123` |
| `service-account list` | `soha service-account list [options]` | List AI Gateway service accounts. | `soha service-account list` |
| `service-account create` | `soha service-account create [options]` | Create a service account. | `soha service-account create --name runner --role-ids operator` |
| `service-account token-list` | `soha service-account token-list [options]` | List service account tokens. | `soha service-account token-list` |
| `service-account token-create` | `soha service-account token-create [options]` | Create a service account token. | `soha service-account token-create --service-account-id sa_123 --name ci` |
| `service-account token-revoke` | `soha service-account token-revoke <token-id> [options]` | Revoke a service account token. | `soha service-account token-revoke sat_123` |
| `audit list` | `soha audit list [options]` | Query AI Gateway audit logs. | `soha audit list --tool-name k8s.pods.logs --limit 20` |
| `approval list` | `soha approval list [options]` | List approval requests. | `soha approval list --status pending` |
| `approval timeline` | `soha approval timeline <approval-id> [options]` | Show an approval timeline. | `soha approval timeline approval_123` |
| `approval approve` | `soha approval approve <approval-id> [options]` | Approve an approval request. | `soha approval approve approval_123 --comment "ship"` |
| `approval reject` | `soha approval reject <approval-id> [options]` | Reject an approval request. | `soha approval reject approval_123 --comment "needs review"` |
| `approval cancel` | `soha approval cancel <approval-id> [options]` | Cancel an approval request. | `soha approval cancel approval_123` |
| `governance status` | `soha governance status [options]` | Show AI Gateway governance health and metrics. | `soha governance status --window-hours 48`<br>`soha governance status --json` |
| `cloud fleet diagnostics` | `soha cloud fleet diagnostics --tenant-id <tenant-id> --fleet-id <fleet-id> [--output summary\|json\|yaml]` | Show Cloud managed agent fleet capability diagnostics. | `soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123`<br>`soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123 --output json` |
| `profile list` | `soha profile list` | List local profiles. | `soha profile list` |
| `profile show` | `soha profile show [profile]` | Show one local profile with tokens redacted. | `soha profile show default` |
| `profile use` | `soha profile use <profile>` | Switch the current profile. | `soha profile use cloud` |
| `context show` | `soha context show [options]` | Show AI client context headers and marketplace defaults. | `soha context show` |
| `context set` | `soha context set [options]` | Update AI client context headers and marketplace defaults. | `soha context set --skill-id k8s-sre --source codex`<br>`soha context set --marketplace https://marketplace.opensoha.com --marketplace-source-id opensoha-official` |
| `mcp` | `soha mcp [options] \| soha mcp <start\|install> [options]` | Run the Soha MCP stdio server or print client configuration. | `soha mcp`<br>`soha mcp --base-url https://soha.example.com` |
| `mcp start` | `soha mcp start [options]` | Compatibility form of soha mcp. | `soha mcp start --profile default` |
| `mcp install` | `soha mcp install [options]` | Print MCP client configuration. | `soha mcp install --command /usr/local/bin/soha`<br>`soha mcp install --base-url https://soha.example.com` |
| `skill list` | `soha skill list [options]` | List skills from a local or verified release source. | `soha skill list` |
| `skill install` | `soha skill install [options] [skill-id...]` | Install skills from a local or verified release source. | `soha skill install --dest ~/.soha/skills k8s-sre` |
| `skill status` | `soha skill status [options]` | Show the active and previous managed skills generations. | `soha skill status`<br>`soha skill status --scope project --json` |
| `skill update` | `soha skill update [options] [skill-id...]` | Update managed skills from the latest stable release. | `soha skill update`<br>`soha skill update --all` |
| `skill remove` | `soha skill remove [options] <skill-id...>` | Remove managed skills with a rollback generation. | `soha skill remove k8s-sre` |
| `skill rollback` | `soha skill rollback [options]` | Switch back to the previous verified skills generation. | `soha skill rollback` |
| `setup` | `soha setup [target] [--client target] [--mode mcp\|skill\|both] [--scope user\|project] [options]` | Configure Soha MCP and skills for an AI agent or IDE. | `soha setup --client codex`<br>`soha setup --client codex --scope project`<br>`soha setup --client claude --mode mcp --base-url https://soha.example.com`<br>`soha setup --client codex --check` |
| `add` | `soha add [target] [options]` | Compatibility alias for configuring Soha MCP and skills. | `soha add --profile local --server http://localhost:8080`<br>`soha add claude --profile default --server https://mcp.opensoha.com --command /usr/local/bin/soha` |
| `plugin search` | `soha plugin search [options]` | Search plugin marketplace entries. | `soha plugin search --query k8s`<br>`soha plugin search --marketplace https://marketplace.opensoha.com --source-id opensoha-official` |
| `plugin show` | `soha plugin show [options] <plugin-id>` | Show marketplace, installed, or manifest details. | `soha plugin show --installed opensoha.k8s-sre-pack`<br>`soha plugin show --marketplace https://marketplace.opensoha.com --version 0.1.0 opensoha.k8s-sre-pack` |
| `plugin install` | `soha plugin install [options] [plugin-id]` | Install a marketplace or manifest plugin. | `soha plugin install --enable opensoha.k8s-sre-pack`<br>`soha plugin install --marketplace https://marketplace.opensoha.com --version 0.1.0 opensoha.k8s-sre-pack` |
| `plugin list` | `soha plugin list [options]` | List installed plugins. | `soha plugin list` |
| `plugin enable` | `soha plugin enable <plugin-id> [options]` | Enable an installed plugin. | `soha plugin enable opensoha.k8s-sre-pack` |
| `plugin disable` | `soha plugin disable <plugin-id> [options]` | Disable an installed plugin. | `soha plugin disable opensoha.k8s-sre-pack` |
| `plugin upgrade` | `soha plugin upgrade [options] <plugin-id>` | Upgrade an installed plugin. | `soha plugin upgrade opensoha.k8s-sre-pack`<br>`soha plugin upgrade --marketplace https://marketplace.opensoha.com --version 0.2.0 opensoha.k8s-sre-pack` |
| `plugin config` | `soha plugin config <plugin-id> [options]` | Configure metadata and secret refs. | `soha plugin config opensoha.k8s-sre-pack --secret-ref token=secret://k8s/token` |
| `plugin remove` | `soha plugin remove <plugin-id> [options]` | Remove an installed plugin. | `soha plugin remove opensoha.k8s-sre-pack` |
| `diagnose` | `soha diagnose [--client <name>] [--output text\|json] [options]` | Check profile, client setup, and Gateway connectivity. | `soha diagnose --client codex --output json`<br>`soha diagnose --tool k8s.pods.logs --resource soha://k8s/runtime` |
| `completion` | `soha completion [bash\|zsh\|fish\|powershell]` | Print shell completion script. | `soha completion bash`<br>`soha completion zsh`<br>`soha completion fish`<br>`soha completion powershell` |
| `docs` | `soha docs [--format markdown]` | Generate CLI command reference documentation. | `soha docs --format markdown` |
