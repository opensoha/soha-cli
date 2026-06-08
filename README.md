# OpenSoha Soha CLI

This repository contains the standalone command-line client for OpenSoha Soha.
The CLI command is `soha`, and the Go module is `github.com/opensoha/soha-cli`.

The CLI connects to a Soha HTTP API by `base_url`, so the same client can talk
to a self-hosted Soha server or a compatible Soha Cloud endpoint. Cloud-specific
lifecycle, billing, quota, and operations logic belongs outside this repository.

## Build

```sh
go build ./cmd/soha
```

## Usage

```sh
soha login --server http://localhost:8080 --login ada
soha login --server https://cloud.soha.run --login ada
soha capabilities
soha mcp install
```

Local profiles are stored under `~/.soha/config.json` by default. Set
`SOHA_CONFIG` to use another config file, `SOHA_SERVER` to override the server
URL, or `SOHA_TOKEN` to override the access token for a single invocation.

## License

This repository is licensed under the Apache License 2.0. See
[LICENSE](./LICENSE) for the full license text.
