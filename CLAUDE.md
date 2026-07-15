# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
# MCP server binary
go build -o pikpakdriver-mcp ./mcp/

# CLI binary
go build -o pikpakdriver ./cli/

# All tests
go test ./pkg/... ./cli/...

# Single package
go test ./pkg/pikpak/

# Dependencies
go mod tidy
```

> **Note**: This project has no tests yet — `go test` will find nothing until they're added.

## Project Overview

A Go CLI + MCP Server for [PikPak](https://mypikpak.com/) cloud storage (like 115driver, but for PikPak). Two entry points share a common client package.

## Architecture

```
mcp/main.go                  # MCP server entry (also has --email/--password login mode)
├── mcp/server/
│   ├── server.go            # Registers 5 tool groups, starts stdio transport
│   └── tools/               # 28 MCP tools (Account, File, Offline, Share, Events)
cli/main.go                  # CLI entry (uses Cobra)
├── cli/cmd/                 # 19 subcommands organized by feature area
│   (no internal packages)
pkg/pikpak/
├── client.go                # HTTP client: auth flow, token refresh, all API methods
├── models.go                # Data models + credential/session file I/O
```

**Key design rule**: `pkg/pikpak/` is the shared layer — CLI and MCP server both call the same `Client` methods. Never add CLI-specific or MCP-specific logic to `pkg/pikpak/`.

**Two login paths**:
- MCP binary (`main.go`): accepts `--email --password [--proxy]` flags for first-time setup, then runs as a stdio-based MCP server
- CLI binary (`cli/main.go`): uses Cobra; `pikpakdriver login --email --password` subcommand, everything else via subcommands

**Codegraph index** is initialized — use `codegraph_explore` to navigate the codebase efficiently.

## Client Lifecycle

- `pikpak.NewClient(cliProxy)` — auto-login from saved credentials/session. Proxy priority: `--proxy` flag > `config.json proxy` > none.
- `pikpak.NewLoginClient(email, proxyURL)` — bare client for first-time login only.
- `client.Login(email, password)` returns `*CaptchaInitResponse` if CAPTCHA is required (user must open the URL in a browser to resolve).
- `client.AccessToken()` — lazy refresh if token expired.
- `doRequest()` auto-retries with refreshed token on 401.

All session/credential files live in `~/.config/pikpakdriver/`:
- `config.json` — email, password, proxy
- `session.json` — access_token, refresh_token, expires_at_unix, device_id

Atomic writes via `write .tmp → rename`.

## CLI Pattern (Cobra)

Each subcommand is in `cli/cmd/<name>.go`. The pattern:

1. Define a `var xCmd = &cobra.Command{Use, Short, Args, RunE}`.
2. `RunE` calls `client.<Method>()`, checks error, formats output.
3. If `jsonOutput` global flag is set, call `printJSON()`; else print human-readable.
4. Register in `init()`: `rootCmd.AddCommand(xCmd)`.

Shared globals in `root.go`: `client`, `proxyFlag`, `jsonOutput`, `emailFlag`, `passwordFlag`.

`PersistentPreRunE` in rootCmd auto-creates the client for all commands except `login`, `help`, `completion`, `version`.

## MCP Tool Pattern

Each tool group has its own file under `mcp/server/tools/`. The pattern:

1. Define a struct holding `*pikpak.Client`.
2. Define typed args structs with `json` and `jsonschema` tags.
3. `RegisterTools(server)` calls `mcp.AddTool(server, &mcp.Tool{Name, Description}, handler)`.
4. Handler calls `client.<Method>()`, returns `errorResult` or `jsonResult`/`successResult`.

Helper functions: `errorResult()`, `successResult()`, `jsonResult()`.

## Region Detection

PikPak blocks mainland China IPs. `isRegionRestricted()` checks error messages for keywords like "region", "forbidden", "403", "connection reset", etc. The `RegionError` type wraps these for the login flow, which prints a friendly Chinese message suggesting proxy usage.

## Important Dependencies

- `github.com/modelcontextprotocol/go-sdk v1.1.0` — MCP framework
- `github.com/spf13/cobra v1.9.1` — CLI framework
- `github.com/spf13/viper v1.21.0` — config reading (CLI auth only; the shared package uses raw JSON)
