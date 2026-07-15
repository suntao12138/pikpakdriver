<p align="right">
  <strong>中文</strong> | <a href="./README.md">English</a>
</p>

# pikpakdriver

> 基于 Go 的 [PikPak](https://mypikpak.com/) 云盘 CLI 和 MCP 服务端 —— 独立于 Rust 版 [`pikpaktui`](https://github.com/niuhuan/pikpak-tui) 项目。

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## 功能特性

- **CLI 客户端** (`pikpakdriver`) — 19 个命令，覆盖 PikPak 全部操作
- **MCP 服务端** (`pikpakdriver-mcp`) — 28 个 MCP 工具，支持 AI 智能体集成（Hermes、Claude 等）
- **全 API 覆盖** — 文件、回收站、离线下载、分享链接、事件、账户
- **代理支持** — CLI 参数 & 配置文件；优先级：`--proxy` > `config.json` > 不用代理
- **自动登录** — 凭证保存到 `config.json`，session 自动续期
- **完全独立** — 独立配置和 session 文件在 `~/.config/pikpakdriver/`，不依赖 `pikpaktui`

---

## 快速开始

### 1. 安装

```bash
# 下载二进制（或从源码编译）
# 预编译二进制见 Releases 页面

# 从源码编译：
git clone https://github.com/suntao12138/pikpakdriver.git
cd pikpakdriver
go build -o pikpakdriver-mcp .           # MCP 服务端
go build -o pikpakdriver ./cli/          # CLI 客户端
```

### 2. 登录（一次性的）

> ⚠️ PikPak 屏蔽中国大陆 IP，需要使用代理。

```bash
# 登录（带代理）
./pikpakdriver login --email your@email.com --password yourpass --proxy http://127.0.0.1:7890

# 验证
./pikpakdriver --proxy http://127.0.0.1:7890 whoami
```

凭证和 session 保存在 `~/.config/pikpakdriver/`，后续使用无需重新登录。

### 3. 部署到 PATH

```bash
mv pikpakdriver-mcp ~/.local/bin/
mv pikpakdriver ~/.local/bin/
```

---

## CLI 使用

### 账户

```bash
pikpakdriver whoami          # 账户信息（存储、VIP）
pikpakdriver login           # 邮箱密码登录
```

### 文件操作

```bash
pikpakdriver ls [parent_id]           # 列出文件
pikpakdriver info <file_id>           # 文件详情
pikpakdriver mkdir <parent_id> <name> # 创建文件夹
pikpakdriver rename <id> <new_name>   # 重命名
pikpakdriver mv <id> <target_id>      # 移动
pikpakdriver cp <id> <target_id>      # 复制
pikpakdriver rm <id...>               # 移到回收站（-P 永久删除）
pikpakdriver link <file_id>           # 获取下载链接
```

### 标星

```bash
pikpakdriver star <id...>     # 标星文件
pikpakdriver unstar <id...>   # 取消标星
pikpakdriver starred [limit]  # 列出已标星文件
```

### 回收站

```bash
pikpakdriver trash ls              # 列出回收站
pikpakdriver trash restore <id...> # 从回收站恢复
pikpakdriver trash empty           # 清空回收站
```

### 离线下载

```bash
pikpakdriver offline add <磁力链接|URL>  # 添加离线任务
pikpakdriver offline ls                 # 列出任务
pikpakdriver offline info <task_id>     # 任务详情
pikpakdriver offline rm <task_id>       # 删除任务
pikpakdriver offline retry <task_id>    # 重试失败任务
```

### 分享链接

```bash
pikpakdriver share create <file_id...>   # 创建分享
pikpakdriver share ls                    # 分享列表
pikpakdriver share rm <share_id...>      # 删除分享
pikpakdriver share info <share_id>       # 分享信息
pikpakdriver share save <share_id> <to>  # 保存他人分享
```

### 事件 & 版本

```bash
pikpakdriver events [limit]    # 最近操作事件
pikpakdriver version            # 版本信息
```

### 全局标志

| 标志 | 说明 |
|------|------|
| `--proxy <url>` | HTTP 代理（如 `http://127.0.0.1:7890`） |
| `-j, --json` | JSON 格式输出 |
| `-h, --help` | 帮助 |

---

## MCP 服务端

MCP 服务端提供 28 个工具，供 AI 智能体集成。配置到 Hermes Agent：

```yaml
# ~/.hermes/config.yaml
mcp_servers:
  pikpakdriver:
    enabled: true
    command: /home/suntao/.local/bin/pikpakdriver-mcp
    args: []
```

### 可用工具

| 类别 | 工具 |
|------|------|
| **账户** | `getAccountInfo` |
| **文件** | `listFiles` `getFileInfo` `getDownloadLink` `mkdir` `rename` `moveFiles` `copyFiles` `starFiles` `unstarFiles` `listStarred` |
| **回收站** | `trashFiles` `untrashFiles` `listTrash` `emptyTrash` `deleteFiles` |
| **离线** | `addOfflineTask` `listOfflineTasks` `getOfflineTask` `deleteOfflineTask` `retryOfflineTask` |
| **分享** | `createShare` `listShares` `deleteShares` `saveShare` `getShareInfo` `shareDetail` |
| **事件** | `listEvents` |

---

## 配置

所有配置文件存储在 `~/.config/pikpakdriver/`：

```
~/.config/pikpakdriver/
├── config.json        # 邮箱、密码、代理
└── session.json       # access_token、refresh_token（自动维护）
```

### 代理优先级

1. `--proxy` CLI 参数（最高）
2. `config.json` 中的 `proxy` 字段
3. 不用代理（默认）

---

## 项目架构

```
pikpakdriver/
├── main.go                    # MCP 服务端入口
├── cli/
│   ├── main.go                # CLI 入口
│   ├── cmd/                   # 19 个子命令（cobra）
│   └── internal/auth/         # 凭证读取
├── mcp/server/
│   ├── server.go              # MCP 注册
│   └── tools/                 # 28 个 MCP 工具实现
├── pkg/pikpak/
│   ├── models.go              # 数据模型（共享）
│   └── client.go              # HTTP 客户端（共享）
└── go.mod
```

`pkg/pikpak/` 包由 CLI 和 MCP 服务端共享 —— 所有 API 调用经过同一客户端层。

---

## 开发

```bash
# 前置要求
go 1.23+

# 完整编译
cd ~/Tools_Pro/pikpakdriver
go build -o pikpakdriver-mcp .
go build -o pikpakdriver ./cli/

# 运行测试
go test ./pkg/... ./cli/...
```

完整测试报告见 [TEST_REPORT.md](./TEST_REPORT.md)。

---

## 相关项目

- [115driver](https://github.com/SheltonZhu/115driver) — 115 网盘 Go SDK/CLI/MCP（本项目的架构参考）
- [pikpaktui](https://github.com/niuhuan/pikpak-tui) — 原始的 Rust TUI 版 PikPak 客户端（完全独立）

---

## 许可证

MIT
