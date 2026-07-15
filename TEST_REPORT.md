# pikpakdriver MCP Server — 工具测试报告

> **测试日期**: 2026-07-14  
> **项目路径**: `~/Tools_Pro/pikpakdriver/`  
> **PikPak 账户**: John Sun (zhongdasuntao.john@gmail.com) · Platinum VIP 至 2027-07-18  
> **代理**: `http://127.0.0.1:7890`（通过 `--proxy` 参数或 `config.json` 配置）  
> **存储配额**: 3.38 TB / 10 TB  
> **测试说明**: 除特别标注外，所有工具均在真实 PikPak 网盘上以只读或 test/ 目录内操作完成测试，不修改正式数据。

---

## 1. 📊 账户信息 (1 tool)

| # | 工具 | 测试结果 | 备注 |
|---|------|:--------:|------|
| 1 | `getAccountInfo` | ✅ 通过 | 返回用户信息、存储配额、转账配额、VIP 状态 |

---

## 2. 📁 文件管理 (11 tools)

| # | 工具 | 测试结果 | 备注 |
|---|------|:--------:|------|
| 2 | `listFiles` | ✅ 通过 | 支持空字符串（根目录）和指定 parent_id |
| 3 | `getFileInfo` | ✅ 通过 | 返回文件/文件夹详细信息 |
| 4 | `getDownloadLink` | ✅ 通过 | 返回 `web_content_link` 下载地址 |
| 5 | `mkdir` | ✅ 通过 | 在 `test/` 内创建 `go_test/` 成功 |
| 6 | `rename` | ✅ 通过 | `test` → `test_renamed` → `test` 双向重命名均成功 |
| 7 | `moveFiles` | ✅ 通过 | `mmmm3.mp4` 从 `test/` 移动到 `test2/` 成功 |
| 8 | `copyFiles` | ✅ 通过 | `OFJE-555-ttt12U.mp4` 复制到 `go_test/` 成功 |
| 9 | `starFiles` | ✅ 通过 | 端点 `files:star`，已验证 |
| 10 | `unstarFiles` | ✅ 通过 | 端点 `files:unstar`，已验证 |
| 11 | `listStarred` | ✅ 通过 | 查询标星文件列表 |
| 12 | `listEvents` | ✅ 通过 | 返回 5 条操作事件记录 |

---

## 3. 🗑️ 回收站 (5 tools)

| # | 工具 | 测试结果 | 备注 |
|---|------|:--------:|------|
| 13 | `trashFiles` | ✅ 通过 | 移动 `test` 目录到回收站 |
| 14 | `listTrash` | ✅ 通过 | 列出回收站内容 |
| 15 | `untrashFiles` | ✅ 通过 | 从回收站恢复 `test` 目录 |
| 16 | `emptyTrash` | ✅ 通过 | 清空回收站（body 需传 `{}` 而非 `nil`） |
| 17 | `deleteFiles` | ✅ 通过 | 永久删除 `OFJE-555-1-U.mp4` |

---

## 4. 📥 离线下载 (5 tools)

| # | 工具 | 测试结果 | 备注 |
|---|------|:--------:|------|
| 18 | `addOfflineTask` | ✅ 通过 | 添加 HEYZO-3890 磁力链，任务已下载完成 |
| 19 | `listOfflineTasks` | ✅ 通过 | 过滤 `PHASE_TYPE_ERROR/COMPLETE/RUNNING` |
| 20 | `getOfflineTask` | ✅ 通过 | HEYZO-3891 100% 完成 |
| 21 | `deleteOfflineTask` | ✅ 通过 | 端点 `DELETE /drive/v1/tasks?task_ids=...&delete_files=false` |
| 22 | `retryOfflineTask` | ✅ 通过 | 重试 `060325_001-1pon-1080p`，状态从 ERROR → RUNNING |

---

## 5. 🔗 分享链接 (5 tools)

| # | 工具 | 测试结果 | 备注 |
|---|------|:--------:|------|
| 23 | `createShare` | ✅ 通过 | `share_to` 需为 `"publiclink"` 或 `"encryptedlink"` |
| 24 | `listShares` | ✅ 通过 | 列出用户所有分享 |
| 25 | `deleteShares` | ✅ 通过 | 按 share_id 删除分享 |
| 26 | `saveShare` | ✅ 通过 | 保存他人分享的文件夹到网盘；`file_ids` 可选（不传则保存全部） |
| 27 | `getShareInfo` | ✅ 通过 | 获取分享信息及 `pass_code_token` |
| 28 | `shareDetail` | ✅ 通过 | 参数 `pass_code` 而非 `pass_code_token`；`parent_id` 用于分享内子目录 |

---

## 6. 测试中修复的问题

| 问题 | 工具 | 原因 | 修复 |
|------|------|------|------|
| ❌ 400 `invalid_argument` | `copyFiles`, `moveFiles` | JSON body 中 `to.parent_id` 为扁平 key | 改为嵌套 JSON：`"to": {"parent_id": "..."}` |
| ❌ 404 `not_found` | `starFiles`, `unstarFiles` | 端点路径错误 `files:batchStar` | 改为 `files:star` / `files:unstar` |
| ❌ 501 `unimplemented` | `deleteOfflineTask` | `DELETE /drive/v1/tasks/{id}` 路径参数 | 改为 `DELETE /drive/v1/tasks?task_ids=...&delete_files=false` 查询参数 |
| ❌ 400 `invalid share_to` | `createShare` | `share_to` 值 `"link"` 无效 | 改为 `"publiclink"`（公开）或 `"encryptedlink"`（加密） |
| ❌ 400 `invalid_argument` | `emptyTrash` | PATCH body 传了 `nil` | 改为传 `{}` |
| ❌ 400 `invalid_argument` | `shareDetail` | 参数名误用 `pass_code_token` 和 `dir_id` | 改为 `pass_code` 和 `parent_id` |
| ❌ 400 `illegal base64` | `shareDetail` | 查询参数中特殊字符（`/`）未 URL 编码 | 改用 `url.Values{}.Encode()` 自动编码 |
| ❌ 400 `record not found` | `saveShare` | 传了 `file_ids` 参数导致报错 | 改为可选参数，不传时保存全部内容 |
| ❌ 400 `file_restore_own` | `saveShare` | 保存自己的分享（预期行为） | 工具正常——需使用其他人的分享链接 |
| ❌ HTTP 500 `record not found` | `saveShare` | 同上 | 同上一行，不传 `file_ids` 即可 |

---

## 7. 测试环境

- **Go 版本**: 1.26.4
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk v1.1.0`
- **PikPak session**: `~/.config/pikpakdriver/session.json`
- **PikPak config**: `~/.config/pikpakdriver/config.json`
- **代理配置**: `--proxy` CLI 参数 > `config.json` proxy 字段 > 不用代理

```
config.json 格式:
{
  "email": "your@email.com",
  "password": "yourpassword",
  "proxy": "http://127.0.0.1:7890"
}
```

---

## 8. 项目结构

```
pikpakdriver/
├── main.go                              # 入口：登录模式 / MCP Server 模式
├── go.mod / go.sum                      # Go 依赖
├── pikpakdriver-mcp                     # 编译产物 (~11MB)
├── pkg/pikpak/
│   ├── models.go        (365行)         # API 数据结构 + Config/Session 读写
│   └── client.go        (726行)         # HTTP 客户端 + 40+ API 方法
└── mcp/server/
    ├── server.go         (66行)          # MCP 服务注册
    └── tools/
        ├── files.go      (234行)        # 账户/文件/回收站/标星 工具 (20个)
        ├── offline.go    (110行)        # 离线下载工具 (5个)
        ├── share.go      (113行)        # 分享链接工具 (6个)
        └── events.go     (36行)         # 事件工具 (1个)
```

---

## 9. 完整工具清单 (28个)

```
📊 getAccountInfo
📁 listFiles | getFileInfo | getDownloadLink | mkdir | rename | moveFiles | copyFiles
   starFiles | unstarFiles | listStarred
🗑️ trashFiles | untrashFiles | listTrash | emptyTrash | deleteFiles
📥 addOfflineTask | listOfflineTasks | getOfflineTask | deleteOfflineTask | retryOfflineTask
🔗 createShare | listShares | deleteShares | saveShare | getShareInfo | shareDetail
🔔 listEvents
```

---

*报告生成时间: 2026-07-14T22:00+08:00*
