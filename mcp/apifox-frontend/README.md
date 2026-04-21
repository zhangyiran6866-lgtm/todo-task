# Apifox Frontend MCP

This MCP server synchronizes a clearly scoped Apifox API contract into the frontend TypeScript API layer.

## Purpose

- Fetch OpenAPI data from an Apifox project.
- Filter APIs by an explicit scope such as `module=auth`, `path-prefix=/tasks`, or selected `operationIds`.
- Generate TypeScript request/response types and API request functions.
- Update only frontend API contract files, such as `packages/frontend/src/api/auth.ts`.
- Return a detailed change report.

This MCP is an API contract synchronizer. It does not update Vue views, Pinia stores, router files, or business logic.

## Tool

The server exposes one MCP tool:

| Tool | Description |
|------|-------------|
| `sync_frontend_api_from_apifox` | Fetch a scoped Apifox OpenAPI contract and generate/update frontend API TypeScript code. |

## Input Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `apifoxToken` | Yes | Apifox Personal Access Token. Never persisted. |
| `projectId` | Yes | Apifox Project ID, not Team ID. |
| `scope` | Yes | Explicit sync scope. Supports `module`, `path-prefix`, and `operation-ids`. |
| `frontend` | No | Frontend generation settings. Defaults to `packages/frontend/src/api`, `@/utils/request`, and `/api/v1`. |
| `writeMode` | No | `dry-run` or `write`. Defaults to `dry-run`. |
| `updateStrategy` | No | Currently supports `generated-block`. |
| `options` | No | Generation and safety options. |

Example:

```json
{
  "apifoxToken": "xxx",
  "projectId": "123456",
  "scope": {
    "type": "module",
    "value": "auth"
  },
  "frontend": {
    "apiDir": "packages/frontend/src/api",
    "requestImportPath": "@/utils/request",
    "basePath": "/api/v1"
  },
  "writeMode": "dry-run",
  "updateStrategy": "generated-block",
  "options": {
    "includeTypes": true,
    "includeRequestFunctions": true,
    "allowInsertGeneratedBlock": false,
    "runTypeCheck": false
  }
}
```

## Scope Rules

| Scope Type | Example | Behavior |
|------------|---------|----------|
| `module` | `{ "type": "module", "value": "auth" }` | Uses built-in module mapping. |
| `path-prefix` | `{ "type": "path-prefix", "value": "/tasks" }` | Matches APIs whose normalized frontend path starts with the prefix. |
| `operation-ids` | `{ "type": "operation-ids", "value": ["login", "register"] }` | Matches exact OpenAPI operation IDs. |

Built-in module mapping:

| Module | Path Prefix | Target File |
|--------|-------------|-------------|
| `auth` | `/auth` | `packages/frontend/src/api/auth.ts` |
| `user` | `/users` | `packages/frontend/src/api/user.ts` |
| `task` | `/tasks` | `packages/frontend/src/api/task.ts` |

## Safety Rules

- Default mode is `dry-run`.
- Write mode updates only the generated block:

```ts
// <apifox-generated module="auth">
// ...
// </apifox-generated>
```

- If the target file exists without a generated block, write mode is skipped unless `options.allowInsertGeneratedBlock` is `true`.
- The MCP does not modify Vue pages, Pinia stores, router files, or business logic.
- The MCP does not save or print Apifox tokens.
- Generated code never uses `any`; unknown schemas are emitted as `unknown` and reported as warnings.

## Build

This MCP package is intentionally isolated under `mcp/` and is not part of the root pnpm workspace package list. Install dependencies from this package directory with `--ignore-workspace`:

```bash
cd mcp/apifox-frontend
pnpm install --ignore-workspace
cd ../..
```

```bash
pnpm --dir mcp/apifox-frontend build
```

## Run

```bash
pnpm --dir mcp/apifox-frontend start
```

The server communicates over stdio and is intended to be launched by an MCP-capable client.

## Apifox OpenAPI Export

This MCP calls:

```text
POST https://api.apifox.com/v1/projects/{projectId}/export-openapi
```

with OpenAPI 3.1 JSON export options.

## Notes For AI Agents

- Read the root `AGENTS.md` before using this tool.
- Use a clear scope. Do not run broad frontend syncs without module/path/operation boundaries.
- Keep local-only credentials under the root `auth/` directory; never commit tokens.
- After write mode, run `pnpm --filter frontend build` when `options.runTypeCheck` is not enabled.
