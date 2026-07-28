---
name: soha-cli
description: >-
  Implement or review the standalone `soha` command-line client in
  `cmd/soha/**`, `internal/sohacli/**`, `docs/commands.md`, `Dockerfile`,
  release workflows, and CLI-facing README content. Use when adding commands,
  flags, completion entries, generated command docs, profile or token handling,
  auth refresh behavior, AI Gateway tool/resource/prompt calls, MCP stdio
  support, `soha add` agent/IDE integration, local skill installation,
  plugin marketplace commands, AI platform knowledge/evaluation/runtime
  commands, cloud diagnostics, or CLI packaging. This skill
  enforces stdlib `flag` command patterns, `Runtime` I/O injection, safe local
  profile storage, redacted output, released `soha-contracts` compatibility,
  and no imports from the core `soha` repository internals.
---

# Soha CLI

## Overview

Use this skill for the `soha-cli` repository only. The CLI is a standalone Go
client that talks to Soha APIs through released `github.com/opensoha/soha-contracts`
types and local HTTP calls. It should remain testable without a real server.

## Workflow

1. Read `internal/sohacli/command_metadata.go` before command changes. Command help, docs, completion hints, and dispatch all flow from this metadata.
2. Add command behavior in the smallest existing owner: profiles/context, MCP, skills, plugins, governance, AI platform/knowledge/evaluation, cloud diagnostics, tokens, service accounts, or Gateway calls.
3. Keep the `Run(ctx, args, Runtime)` boundary. Use `Runtime.In`, `Runtime.Out`, `Runtime.Err`, `Runtime.ConfigPath`, and injectable HTTP clients in tests; do not write directly to `os.Stdout` or `os.Stderr` outside `cmd/soha/main.go`.
4. Use stdlib `flag` plus `newRuntimeFlagSet`. Return errors from handlers and let `Run` map them to exit codes.
5. Use `APIClient`, the focused AI platform client, and contracts DTOs for HTTP work. Do not import `github.com/opensoha/soha/internal/**` or sibling checkout internals.
6. Keep the committed `go.mod` on a released `soha-contracts` tag. Use temporary local `go.work` only for unreleased multi-repo contract work.
7. When command surface changes, update `topLevelCommandSpecs`, completion words where relevant, and regenerate `docs/commands.md` with `go run ./cmd/soha docs --format markdown > docs/commands.md`.

## Security Rules

- Local config defaults to `~/.soha/config.json` or `SOHA_CONFIG`; preserve `0700` parent dirs and `0600` config file permissions.
- Never print full access tokens, refresh tokens, passwords, service account tokens, or plugin secrets. Use existing redaction helpers and test for leaks.
- `SOHA_TOKEN` is a one-command override and must not be persisted back to the profile.
- Expired profile access tokens should refresh with the stored refresh token and persist rotated token state; refresh requests must not send an Authorization header.
- Interactive passwords go to stderr prompts and must not echo into stdout, logs, or test failures.
- `soha add` and skill/plugin install paths must stay idempotent, path-safe, and dry-run friendly.

## Testing

- For command logic, write tests through `Run` with injected buffers, temp config paths, and `httptest.Server`.
- For HTTP clients, cover request path, method, auth headers, refresh behavior, and redacted output.
- Run `go test ./...` for normal changes.
- Run `GOWORK=off go test ./...` before changes that touch contracts, module state, Dockerfile, or release behavior.
- Run `go vet ./...` and `go test -race ./...` for concurrency, token refresh, MCP stdio, or release-sensitive work.
