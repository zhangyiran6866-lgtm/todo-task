# Apifox Backend MCP

This MCP server synchronizes the Go backend Swagger/OpenAPI document to an Apifox project.

## Purpose

- Generate the latest Swagger files from `packages/backend`.
- Read `packages/backend/docs/swagger.json`.
- Import the generated OpenAPI document into Apifox.

## Tool

The server exposes one MCP tool:

| Tool | Description |
|------|-------------|
| `sync_apifox` | Generate backend Swagger docs and upload them to Apifox. |

Input parameters:

| Parameter | Required | Description |
|-----------|----------|-------------|
| `apifoxToken` | Yes | Apifox Personal Access Token. |
| `projectId` | Yes | Apifox Project ID, not Team ID. |
| `options` | No | Optional sync options. |
| `options.tls.caCertPath` | No | CA certificate PEM path for strict TLS in custom network environments. |
| `options.tls.allowInsecureTLS` | No | Temporary local fallback only. Disables TLS certificate verification. |

## Local Credentials

Do not store tokens, account passwords, or API keys in this MCP package.

For this repository, keep local-only credentials under the root `auth/` directory:

```text
auth/apifox.md
```

The root `.gitignore` ignores `auth/`, so files in that directory should not be committed.

Recommended fields:

```text
Project name:
Project ID:
Personal Access Token:
```

## Build

```bash
pnpm --dir mcp/apifox-backend build
```

## Run

```bash
pnpm --dir mcp/apifox-backend start
```

The server communicates over stdio and is intended to be launched by an MCP-capable client.

## Backend Swagger Generation

The `sync_apifox` tool runs:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go --parseDependency --parseInternal
```

from:

```text
packages/backend
```

Generated files:

```text
packages/backend/docs/docs.go
packages/backend/docs/swagger.json
packages/backend/docs/swagger.yaml
```

## Notes For AI Agents

- Read the root `AGENTS.md` before using this tool.
- Phase 2 is complete, so this MCP can be used for API documentation synchronization when needed.
- Never print or commit Apifox tokens.
- Prefer strict TLS with `options.tls.caCertPath`; avoid `allowInsecureTLS=true` except temporary local troubleshooting.
- If Swagger generation fails on handler annotations such as `model.Task` or `model.User`, ensure the handler file imports the referenced package, often as a blank import for docs parsing.
- After Swagger generation, run `go test ./...` in `packages/backend` to confirm the backend still compiles.
