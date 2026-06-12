# OpenSoha Soha CLI

This repository contains the standalone command-line client for OpenSoha Soha.
The CLI command is `soha`, and the Go module is `github.com/opensoha/soha-cli`.

The CLI connects to a Soha HTTP API by base URL, so the same client can talk to
a self-hosted Soha server or a compatible Soha Cloud endpoint. Cloud-specific
lifecycle, billing, quota, and operations logic belongs outside this repository.

## Install

Install a tagged module with Go:

```sh
go install github.com/opensoha/soha-cli/cmd/soha@v0.1.0
soha version
```

For release binaries, download the archive for your platform from the GitHub
release, then verify it against `checksums.txt` before placing `soha` on your
`PATH`:

```sh
sha256sum -c checksums.txt --ignore-missing
tar -xzf soha_0.1.0_linux_amd64.tar.gz
install -m 0755 soha /usr/local/bin/soha
```

Windows releases are published as `.zip` archives and use the same
`checksums.txt` file. Release builds inject `version`, `commit`, and `date`
through Go ldflags; `soha version --json` is the machine-readable verification
command for installed binaries.

## Build

```sh
go test ./...
GOWORK=off go test ./...
go build ./cmd/soha
```

`soha-cli` consumes released Go contracts from
`github.com/opensoha/soha-contracts`. A clean checkout should build without a
sibling `../soha-contracts` directory. When developing unreleased contract
changes locally, create a temporary workspace outside the committed module
state:

```sh
go work init . ../soha-contracts
go test ./...
```

Do not commit local `go.work` files unless the release process explicitly asks
for a multi-repository workspace.

## Contracts Compatibility

`soha-cli` treats `github.com/opensoha/soha-contracts` as its public API source.
The committed `go.mod` must point at a released contracts tag, not a sibling
checkout. Contract upgrades should be reviewed with this minimum matrix:

| soha-cli | soha-contracts | Compatible Soha API | Notes |
| --- | --- | --- | --- |
| `v0.1.x` | `v0.1.x` | `v0.1.x` beta API | Supported beta line. |
| `main` | latest released `v0.1.x` unless intentionally bumped | local development API | Must pass `go test ./...` and `GOWORK=off go test ./...`. |

When a contracts change is not backward compatible, release a new contracts tag
first, update this module in a dedicated commit, and run the consumer matrix for
`soha`, `soha-web`, `soha-cli`, and `soha-agent` before tagging the CLI.

## Configuration

Local profiles are stored under `~/.soha/config.json` by default with `0600`
permissions. The file contains access and refresh tokens; keep it out of source
control and shared logs.

Profile `expiresAt` is the access token expiration time. Commands automatically
refresh an expired access token with the stored refresh token and persist the
rotated token set. If a profile has an expired access token and no refresh
token, run `soha login` again. `SOHA_TOKEN` is a one-command access token
override and is never written back to the profile.

| Name | Purpose |
| --- | --- |
| `SOHA_CONFIG` | Use another config file path. |
| `SOHA_SERVER` | Override the configured server URL for a single invocation, or provide the default for `login`; otherwise new profiles default to `https://mcp.opensoha.com`. |
| `SOHA_TOKEN` | Override the configured access token for a single invocation. |
| `SOHA_LOGIN` | Provide the default login name for `login`. |
| `SOHA_PASSWORD` | Provide the login password non-interactively. Prefer short-lived shell scope or a secret manager. |
| `SOHA_HTTP_TIMEOUT` | HTTP request timeout, such as `10s`, `1m`, or bare seconds like `15`. Defaults to `30s`. |
| `SOHA_AI_CLIENT_ID` | Default AI client ID stored by `login`. |
| `SOHA_AI_CLIENT` | Default AI client display name stored by `login`. |

All HTTP commands also accept a global `--timeout` flag. It can be placed before
or after the command:

```sh
soha --timeout 10s capabilities
soha capabilities --timeout 10s
```

Interactive `soha login` reads the password with terminal echo disabled. In
non-interactive use, pass `--password`, set `SOHA_PASSWORD`, or pipe a single
line to stdin.

## Core Workflows

```sh
soha login --server http://localhost:8080 --login ada
soha login --server https://cloud.soha.run --login ada --profile cloud
soha profile list
soha context set --ai-client-id codex-local --ai-client Codex --source codex
soha capabilities --output names
soha diagnose --tool k8s.pods.logs
soha mcp install --profile default --command /usr/local/bin/soha
soha add codex --profile default --command /usr/local/bin/soha --source ../soha-skills
```

### Codex Integration

`soha add codex` writes a `[mcp_servers.soha]` entry to Codex
`config.toml`, installs a Codex-compatible `soha` skill package under
`~/.agents/skills/soha`, and installs the raw Soha runtime skills under
`~/.soha/skills`:

```sh
soha add codex \
  --profile local \
  --server http://localhost:8080 \
  --command "$(command -v soha)" \
  --source ../soha-skills
```

Run `soha add` without a target to choose one or more AI agents and IDEs
interactively:

```sh
soha add --profile local --server http://localhost:8080 --source ../soha-skills
```

The selection accepts numbers, names, or `all`. Supported targets are `codex`,
`claude`, `cursor`, `kiro`, `gemini`, `antigravity`, `antigravity-ide`, and
`trae`.

The command is idempotent for the MCP section and generated skill package. Pass
`--server` or `--base-url` to set the Soha API base URL for the generated
profile. For new profiles, the default API base URL is
`https://mcp.opensoha.com`. Pass `--codex-config`, `--dest`, or
`--runtime-skill-dest` to target non-default locations. Restart Codex or open a
new Codex session after running it so MCP servers and skills are reloaded.

## Command Matrix

For the generated command reference, run `soha docs --format markdown` or read
[`docs/commands.md`](./docs/commands.md).

| Command | Purpose | Common examples |
| --- | --- | --- |
| `version` | Print build version information. | `soha version`, `soha version --json` |
| `login` | Authenticate and store a local profile. | `soha login --server http://localhost:8080 --login ada` |
| `capabilities` | Print the AI Gateway manifest. | `soha capabilities --output names`, `soha capabilities --output inputs` |
| `tool call` | Invoke an AI Gateway tool with JSON input. | `soha tool call k8s.pods.list --input-json '{"clusterId":"local"}'` |
| `resource read` | Read an AI Gateway MCP resource. | `soha resource read soha://k8s/runtime --context-json '{"clusterId":"local"}'` |
| `prompt get` | Get an AI Gateway MCP prompt. | `soha prompt get soha.k8s.diagnose_workload --arguments-json '{}'` |
| `token list` | List personal access tokens. | `soha token list` |
| `token create` | Create a personal access token. | `soha token create --name local --permission-keys ai.gateway.view` |
| `token revoke` | Revoke a personal access token. | `soha token revoke pat_123` |
| `service-account list` | List AI Gateway service accounts. | `soha service-account list` |
| `service-account create` | Create a service account. | `soha service-account create --name runner --role-ids operator` |
| `service-account token-list` | List service account tokens. | `soha service-account token-list` |
| `service-account token-create` | Create a service account token. | `soha service-account token-create --service-account-id sa_123 --name ci` |
| `service-account token-revoke` | Revoke a service account token. | `soha service-account token-revoke sat_123` |
| `audit list` | Query AI Gateway audit logs. | `soha audit list --tool-name k8s.pods.logs --limit 20` |
| `approval list` | List approval requests. | `soha approval list --status pending` |
| `approval timeline` | Show an approval timeline. | `soha approval timeline approval_123` |
| `approval approve` | Approve an approval request. | `soha approval approve approval_123 --comment "ship"` |
| `approval reject` | Reject an approval request. | `soha approval reject approval_123 --comment "needs review"` |
| `approval cancel` | Cancel an approval request. | `soha approval cancel approval_123` |
| `governance status` | Show AI Gateway governance health and metrics. | `soha governance status --window-hours 48`, `soha governance status --json` |
| `cloud fleet diagnostics` | Show Cloud managed agent fleet capability diagnostics. | `soha cloud fleet diagnostics --tenant-id tenant_123 --fleet-id agent-fleet_123` |
| `profile list` | List local profiles. | `soha profile list` |
| `profile show` | Show one local profile with tokens redacted. | `soha profile show default` |
| `profile use` | Switch the current profile. | `soha profile use cloud` |
| `context show` | Show AI client context headers. | `soha context show` |
| `context set` | Update AI client context headers. | `soha context set --skill-id k8s-sre --source codex` |
| `mcp start` | Run the Soha MCP stdio server. | `soha mcp start --profile default` |
| `mcp install` | Print MCP client configuration. | `soha mcp install --command /usr/local/bin/soha` |
| `skill list` | List local Soha skill files. | `soha skill list --source ../soha-skills` |
| `skill install` | Install local Soha skill files. | `soha skill install --source ../soha-skills --dest ~/.soha/skills k8s-sre` |
| `add` | Interactively choose AI agents or IDEs and add Soha MCP plus skills. | `soha add --profile local --server http://localhost:8080 --source ../soha-skills` |
| `add <target>` | Add Soha MCP and a skill package for one target. | `soha add claude --profile default --server https://mcp.opensoha.com --command /usr/local/bin/soha --source ../soha-skills` |
| `plugin search` | Search plugin marketplace entries. | `soha plugin search --query k8s` |
| `plugin show` | Show marketplace, installed, or manifest details. | `soha plugin show opensoha.k8s-sre-pack --installed` |
| `plugin install` | Install a marketplace or manifest plugin. | `soha plugin install opensoha.k8s-sre-pack --enable` |
| `plugin list` | List installed plugins. | `soha plugin list` |
| `plugin enable` | Enable an installed plugin. | `soha plugin enable opensoha.k8s-sre-pack` |
| `plugin disable` | Disable an installed plugin. | `soha plugin disable opensoha.k8s-sre-pack` |
| `plugin upgrade` | Upgrade an installed plugin. | `soha plugin upgrade opensoha.k8s-sre-pack` |
| `plugin config` | Configure metadata and secret refs. | `soha plugin config opensoha.k8s-sre-pack --secret-ref token=secret://k8s/token` |
| `plugin remove` | Remove an installed plugin. | `soha plugin remove opensoha.k8s-sre-pack` |
| `diagnose` | Check profile and Gateway visibility. | `soha diagnose --tool k8s.pods.logs --resource soha://k8s/runtime` |
| `completion` | Print shell completion script. | `soha completion bash`, `soha completion zsh` |
| `docs` | Generate CLI command reference documentation. | `soha docs --format markdown` |

## License

This repository is licensed under the Apache License 2.0. See
[LICENSE](./LICENSE) for the full license text.
