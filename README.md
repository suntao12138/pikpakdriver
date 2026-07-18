<p align="right">
  <a href="./README_cn-ZH.md">中文</a> | <strong>English</strong>
</p>

# pikpakdriver

> A Go-based CLI and MCP Server for [PikPak](https://mypikpak.com/) cloud storage — independent from the Rust-based [`pikpaktui`](https://github.com/niuhuan/pikpak-tui) project.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Features

- **CLI Client** (`pikpakdriver`) — 19 commands covering all PikPak operations
- **MCP Server** (`pikpakdriver-mcp`) — 28 MCP tools for AI agent integration (Hermes, Claude, etc.)
- **Full API coverage** — files, trash, offline downloads, share links, events, account
- **Proxy support** — CLI flag & config file; priority: `--proxy` > `config.json` > no proxy
- **Auto-login** — credentials saved to `config.json`, session auto-refreshed
- **Independent** — own config & session at `~/.config/pikpakdriver/`, no dependency on `pikpaktui`

---

## Quick Start

### 1. Install

```bash
# Download the binary (or build from source)
# Pre-built binaries can be found on the Releases page

# Or build from source:
git clone https://github.com/suntao12138/pikpakdriver.git
cd pikpakdriver
go build -o pikpakdriver-mcp ./mcp/      # MCP Server
go build -o pikpakdriver ./cli/          # CLI Client
```

### 2. Login (one-time)

> ⚠️ PikPak blocks mainland China IPs. Use `--proxy` if needed.

Either method works — pick one:

**Option A: CLI Client**
```bash
./pikpakdriver login --email your@email.com --password yourpass --proxy http://127.0.0.1:7890
```

**Option B: MCP Server** (also supports one-shot login)
```bash
./pikpakdriver-mcp --email your@email.com --password yourpass --proxy http://127.0.0.1:7890
```

Credentials and session are saved to `~/.config/pikpakdriver/`. Subsequent runs do not require re-login.

### 3. Deploy to PATH

```bash
mv pikpakdriver-mcp ~/.local/bin/
mv pikpakdriver ~/.local/bin/
```

---

## CLI Usage

### Account

```bash
pikpakdriver whoami          # Account info (storage, VIP)
pikpakdriver login           # Login with email/password
```

### File Operations

```bash
pikpakdriver ls [parent_id]           # List files
pikpakdriver info <file_id>           # File details
pikpakdriver mkdir <parent_id> <name> # Create folder
pikpakdriver rename <id> <new_name>   # Rename
pikpakdriver mv <id> <target_id>      # Move
pikpakdriver cp <id> <target_id>      # Copy
pikpakdriver rm <id...>               # Trash (use -P for permanent)
pikpakdriver link <file_id>           # Get download URL
```

### Star

```bash
pikpakdriver star <id...>     # Star files
pikpakdriver unstar <id...>   # Unstar files
pikpakdriver starred [limit]  # List starred files
```

### Trash

```bash
pikpakdriver trash ls              # List trash
pikpakdriver trash restore <id...> # Restore from trash
pikpakdriver trash empty           # Empty trash
```

### Offline Downloads

```bash
pikpakdriver offline add <magnet|url>  # Add download task
pikpakdriver offline ls                # List tasks
pikpakdriver offline info <task_id>    # Task details
pikpakdriver offline rm <task_id>      # Delete task
pikpakdriver offline retry <task_id>   # Retry failed task
```

### Share Links

```bash
pikpakdriver share create <file_id...>   # Create share link
pikpakdriver share ls                    # List shares
pikpakdriver share rm <share_id...>      # Delete shares
pikpakdriver share info <share_id>       # Get share info
pikpakdriver share save <share_id> <to>  # Save shared files
```

### Events & Version

```bash
pikpakdriver events [limit]    # Recent file events
pikpakdriver version            # Version info
```

### Global Flags

| Flag | Description |
|------|-------------|
| `--proxy <url>` | HTTP proxy (e.g. `http://127.0.0.1:7890`) |
| `-j, --json` | JSON output format |
| `-h, --help` | Help |

---

## MCP Server

The MCP server supports two modes:

### Mode 1: First-time Login

```bash
./pikpakdriver-mcp --email your@email.com --password yourpass --proxy http://127.0.0.1:7890
```

Credentials are saved to `~/.config/pikpakdriver/` and the program exits. Subsequent runs can start the server directly.

> ⚠️ If CAPTCHA verification is triggered, the command will print a verification URL — open it in a browser to complete the challenge, then retry.

### Mode 2: Start MCP Server

```bash
./pikpakdriver-mcp                                  # Uses proxy from config.json if set
./pikpakdriver-mcp --proxy http://127.0.0.1:7890    # Override proxy
```

Auto-logs in from saved credentials and starts a stdio-based MCP server with 28 tools. Configure it in Claude Desktop / Hermes Agent / any MCP client:

```yaml
# ~/.hermes/config.yaml
mcp_servers:
  pikpakdriver:
    enabled: true
    command: /home/suntao/.local/bin/pikpakdriver-mcp
    args: []
```

### Available Tools

| Category | Tools |
|----------|-------|
| **Account** | `getAccountInfo` |
| **Files** | `listFiles` `getFileInfo` `getDownloadLink` `mkdir` `rename` `moveFiles` `copyFiles` `starFiles` `unstarFiles` `listStarred` |
| **Trash** | `trashFiles` `untrashFiles` `listTrash` `emptyTrash` `deleteFiles` |
| **Offline** | `addOfflineTask` `listOfflineTasks` `getOfflineTask` `deleteOfflineTask` `retryOfflineTask` |
| **Share** | `createShare` `getShareInfo` `saveShare` `shareDetail` `listShares` `deleteShares` |
| **Events** | `listEvents` |

---

## Configuration

All configuration files are stored at `~/.config/pikpakdriver/`:

```
~/.config/pikpakdriver/
├── config.json        # email, password, proxy
└── session.json       # access_token, refresh_token (auto-managed)
```

### Proxy Priority

1. `--proxy` CLI flag (highest)
2. `proxy` field in `config.json`
3. No proxy (default)

---

## Architecture

```
pikpakdriver/
├── mcp/
│   ├── main.go                # Entry point (login mode + MCP server mode)
│   └── server/
│       ├── server.go          # MCP registration
│       └── tools/             # 28 MCP tools (one file per group)
│           ├── account.go
│           ├── events.go
│           ├── files.go
│           ├── offline.go
│           └── share.go
├── cli/
│   ├── main.go                # CLI entry point
│   ├── cmd/                   # 19 subcommands (cobra)
│   └── internal/auth/         # Credential loading
├── pkg/pikpak/
│   ├── models.go              # Data models (shared)
│   └── client.go              # HTTP client (shared)
└── go.mod
```

The `pkg/pikpak/` package is shared between the CLI and MCP server — all API calls go through the same client layer.

---

## Development

```bash
# Prerequisites
go 1.23+

# Full build
cd ~/Tools_Pro/pikpakdriver
go build -o pikpakdriver-mcp ./mcp/
go build -o pikpakdriver ./cli/

# Run tests
go test ./pkg/... ./cli/...
```

See [TEST_REPORT.md](./TEST_REPORT.md) for the complete tool validation report.

---

## Related Projects

- [115driver](https://github.com/SheltonZhu/115driver) — 115 cloud drive Go SDK/CLI/MCP (the architecture reference for this project)
- [pikpaktui](https://github.com/niuhuan/pikpak-tui) — Original Rust TUI client for PikPak (independent)

---

## License

MIT
