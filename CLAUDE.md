# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
# All binaries
go build -o pikpakdriver-mcp ./mcp/      # MCP server
go build -o pikpakdriver ./cli/          # CLI

# Tests (no tests exist yet)
go test ./pkg/... ./cli/...

# Single package
go test ./pkg/pikpak/

# Lint & tidy
go vet ./pkg/... ./mcp/... ./cli/...
go mod tidy
```

## Project Overview

A Go CLI + MCP Server for [PikPak](https://mypikpak.com/) cloud storage (like 115driver, but for PikPak). Two entry points share a common client package `pkg/pikpak/`.

## Architecture

```
pkg/pikpak/               ← SHARED LAYER — all API logic lives here
├── client.go             HTTP client, auth flow, token refresh, all API methods
└── models.go             Data models, session/credential file I/O, constants

mcp/main.go               Entry: --email --password for login, then stdio MCP server
└── mcp/server/
    ├── server.go         Registers 5 tool groups, starts stdio transport
    └── tools/            One file per group
        ├── account.go / events.go / files.go / offline.go / share.go

cli/main.go               Entry: Cobra CLI
├── cli/cmd/              19 subcommands, one file each
└── cli/internal/auth/    Credential loading (small helper)
```

**Key design rule**: `pkg/pikpak/` is the shared layer — CLI and MCP server both call the same `Client` methods. Never add CLI-specific or MCP-specific logic to `pkg/pikpak/`.

**Two login paths**:
- MCP binary (`mcp/main.go`): `--email --password [--proxy]` saves credentials and exits; re-run without flags to start the MCP server (auto-login from saved creds)
- CLI binary (`cli/main.go`): `pikpakdriver login --email --password` subcommand, everything else via subcommands

**Codegraph index** is initialized — use `codegraph_explore` to navigate the codebase efficiently.

## Client Lifecycle

```go
pikpak.NewClient(cliProxy)        // Auto-login: session → refresh → credentials → autoLogin
pikpak.NewLoginClient(email, url) // Bare client, first-time login only
client.Login(email, password)     // Returns *CaptchaInitResponse if CAPTCHA required
```

- `client.Login()` returns a `*CaptchaInitResponse` with URL when CAPTCHA is triggered — user must open in browser, then retry.
- `client.AccessToken()` — lazy refresh if expired.
- `doRequest()` auto-retries with refreshed token on 401.
- Proxy priority: `--proxy` flag > `config.json proxy` > none.

**Session/credential files** in `~/.config/pikpakdriver/`:
- `config.json` — email, password, proxy (written atomically: `.tmp` → rename)
- `session.json` — access_token, refresh_token, expires_at_unix, device_id

Both files use atomic writes: `os.WriteFile(path.tmp, data, 0600)` → `os.Rename(tmpPath, path)`.

## HTTP Layer (pkg/pikpak/client.go)

All authenticated requests flow through a two-layer retry system:

1. `doRequest(method, baseURL, path, query, body)` — gets token, sends request, retries once on 401 (refreshes token)
2. `doRequestWithToken(...)` — single request with given token, no retry

Convenience wrappers: `driveGET`, `drivePOST`, `drivePATCH`, `driveDELETE`, `authGET`.
All return raw `[]byte`; callers unmarshal into typed models.

## CLI Pattern (Cobra)

Each subcommand in `cli/cmd/<name>.go`:

1. Define `var xCmd = &cobra.Command{Use, Short, Args, RunE}`
2. `RunE` calls `client.<Method>()`, checks error, formats output
3. If `jsonOutput` global flag is set, call `printJSON()`; else print human-readable
4. Register in `init()`: `rootCmd.AddCommand(xCmd)`

Shared globals in `root.go`: `client`, `proxyFlag`, `jsonOutput`, `emailFlag`, `passwordFlag`.
`PersistentPreRunE` creates `client` for all commands except `login`, `help`, `completion`, `version`.

Helper: `printJSON(v)` in `helpers.go` — marshals with indent, handles error.

## MCP Tool Pattern

Each tool file under `mcp/server/tools/`:

1. Define a struct holding `*pikpak.Client` (e.g. `FileTools struct{ client *pikpak.Client }`)
2. Define args structs with `json` and `jsonschema` tags
3. `RegisterTools(server)` calls `mcp.AddTool(server, &mcp.Tool{Name, Description}, handler)`
4. Handler signature: `func(ctx, *CallToolRequest, args) (*CallToolResult, any, error)`
5. Return helpers: `errorResult(format, args...)`, `successResult(text)`, `jsonResult(v)`

All helpers and `emptyArgs` are defined in the `tools` package (package-level shared).

## Region Detection

PikPak blocks mainland China IPs. `isRegionRestricted()` checks error messages for keywords like "region", "forbidden", "403", "connection reset", etc. The `RegionError` type wraps these for the login flow, which prints a Chinese-language message suggesting proxy usage.

## Important Dependencies

- `github.com/modelcontextprotocol/go-sdk v1.1.0` — MCP framework
- `github.com/spf13/cobra v1.9.1` — CLI framework
- `github.com/spf13/viper v1.21.0` — config reading (CLI auth only; the shared package uses raw JSON)
