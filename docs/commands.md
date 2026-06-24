# Soha CLI Command Reference

Generated with `soha docs --format markdown`.

| Command | Usage | Purpose | Examples |
| --- | --- | --- | --- |
| `version` | `soha version [--json]` | Print build version information. | `soha version`<br>`soha version --json` |
| `login` | `soha login [options]` | Authenticate and store a local profile. | `soha login --server http://localhost:8080 --login ada` |
| `capabilities` | `soha capabilities [--domain gateway\|platform] [--output json\|yaml\|names\|inputs]` | Print AI Gateway or platform capability metadata. | `soha capabilities --output names`<br>`soha capabilities --output inputs` |
| `tool call` | `soha tool call <name> [options]` | Invoke an AI Gateway tool with JSON input. | `soha tool call k8s.pods.list --input-json '{"clusterId":"local"}'` |
| `resource read` | `soha resource read <uri> [options]` | Read an AI Gateway MCP resource. | `soha resource read soha://k8s/runtime --context-json '{"clusterId":"local"}'` |
| `prompt get` | `soha prompt get <name> [options]` | Get an AI Gateway MCP prompt. | `soha prompt get soha.k8s.diagnose_workload --arguments-json '{}'` |
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
| `context show` | `soha context show [options]` | Show AI client context headers. | `soha context show` |
| `context set` | `soha context set [options]` | Update AI client context headers. | `soha context set --skill-id k8s-sre --source codex` |
| `mcp start` | `soha mcp start [options]` | Run the Soha MCP stdio server. | `soha mcp start --profile default` |
| `mcp install` | `soha mcp install [options]` | Print MCP client configuration. | `soha mcp install --command /usr/local/bin/soha` |
| `skill list` | `soha skill list [options]` | List local Soha skill files. | `soha skill list --source ../soha-skills` |
| `skill install` | `soha skill install [options] [skill-id...]` | Install local Soha skill files. | `soha skill install --source ../soha-skills --dest ~/.soha/skills k8s-sre` |
| `add` | `soha add [target] [options]` | Add Soha MCP and skills to an AI agent or IDE. | `soha add --profile local --server http://localhost:8080 --source ../soha-skills`<br>`soha add claude --profile default --server https://mcp.opensoha.com --command /usr/local/bin/soha --source ../soha-skills` |
| `plugin search` | `soha plugin search [options]` | Search plugin marketplace entries. | `soha plugin search --query k8s` |
| `plugin show` | `soha plugin show [plugin-id] [options]` | Show marketplace, installed, or manifest details. | `soha plugin show opensoha.k8s-sre-pack --installed` |
| `plugin install` | `soha plugin install [plugin-id] [options]` | Install a marketplace or manifest plugin. | `soha plugin install opensoha.k8s-sre-pack --enable` |
| `plugin list` | `soha plugin list [options]` | List installed plugins. | `soha plugin list` |
| `plugin enable` | `soha plugin enable <plugin-id> [options]` | Enable an installed plugin. | `soha plugin enable opensoha.k8s-sre-pack` |
| `plugin disable` | `soha plugin disable <plugin-id> [options]` | Disable an installed plugin. | `soha plugin disable opensoha.k8s-sre-pack` |
| `plugin upgrade` | `soha plugin upgrade <plugin-id> [options]` | Upgrade an installed plugin. | `soha plugin upgrade opensoha.k8s-sre-pack` |
| `plugin config` | `soha plugin config <plugin-id> [options]` | Configure metadata and secret refs. | `soha plugin config opensoha.k8s-sre-pack --secret-ref token=secret://k8s/token` |
| `plugin remove` | `soha plugin remove <plugin-id> [options]` | Remove an installed plugin. | `soha plugin remove opensoha.k8s-sre-pack` |
| `diagnose` | `soha diagnose [options]` | Check profile and Gateway connectivity. | `soha diagnose --tool k8s.pods.logs --resource soha://k8s/runtime` |
| `completion` | `soha completion [bash\|zsh]` | Print shell completion script. | `soha completion bash`<br>`soha completion zsh` |
| `docs` | `soha docs [--format markdown]` | Generate CLI command reference documentation. | `soha docs --format markdown` |
