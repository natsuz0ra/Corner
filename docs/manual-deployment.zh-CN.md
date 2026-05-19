# 手动源码部署

本文说明基于源码的开发启动、生产构建、Docker、配置、沙盒行为和运行时数据目录。面向普通用户的推荐安装方式见主 [README](../README.zh-CN.md)。

## 架构

- **生产：** 单个 Go 二进制同时提供 REST/WebSocket，并嵌入 `web/dist` 静态资源。
- **开发：** `npm run dev` 同时启动 Go 服务和 Vite；Vite 将 `/api`、`/ws` 代理到后端端口，默认 `6247`。
- **数据：** SQLite 默认路径为 `~/.slimebot/storage/data.db`。
- **技术栈：** Go 后端、Vue 3 Web 应用、React + Ink CLI。

## 开发启动

默认端口：后端 **6247**，Vite **7391**。

```bash
make deps
npm run dev
```

手动安装依赖：

```bash
npm install
npm install --prefix frontend
npm run dev
```

首次启动会在缺失时创建 `~/.slimebot/config.cfg`。若存在旧的 `~/.slimebot/.env` 且不存在 `config.cfg`，SlimeBot 会复制旧配置到 `config.cfg`，并保留旧文件不动。

## 从源码构建

生产构建：

```bash
npm run build
# 或
make build
```

运行构建后的 Web 服务：

```bash
./slimebot server
```

从源码启动 CLI TUI：

```bash
npm run cli
```

生成本地 Release 压缩包：

```bash
make package
```

Release 压缩包包含 `install.sh` / `install.ps1` 和 `uninstall.sh` / `uninstall.ps1`。安装脚本支持 `SLIMEBOT_INSTALL_DIR` 与 `SLIMEBOT_BIN_DIR`；卸载脚本使用同样的变量，并额外支持 `SLIMEBOT_HOME` 指定用户数据目录。

打包脚本也会生成独立的 `dist/install.sh`、`dist/install.ps1`、`dist/uninstall.sh` 和 `dist/uninstall.ps1` 发布资产。安装脚本在解压后的 Release 目录外执行时，会解析最新 GitHub Release，下载匹配当前平台的压缩包，然后运行压缩包内的安装器；可通过 `SLIMEBOT_VERSION=v1.26.1` 安装指定版本标签，或用 `SLIMEBOT_REPO=owner/repo` 指向 fork。卸载脚本可直接远程执行，不下载 Release 压缩包，而是按本机安装路径删除已安装内容。

更新行为：

- `slimebot update --check` 检查 GitHub 最新稳定 Release。更新源沿用安装逻辑：优先读取 `SLIMEBOT_REPO`，否则使用 `natsuz0ra/SlimeBot`。
- `slimebot update` 会启动独立 helper 进程，下载匹配当前平台的 Release 压缩包，解压并运行包内安装脚本，同时把进度写入 `~/.slimebot/storage/update-status.json`。
- `slimebot update --version vX.Y.Z` 安装指定 Release tag。当前构建版本为 `dev`、空值，或无法解析为版本号时，界面也会展示这条命令作为恢复入口。
- Web 用户服务模式下，helper 会尝试执行 `slimebot service stop`，完成安装后再尝试 `slimebot service start`。如果当前是前台 `slimebot server` 进程且没有安装服务，helper 完成后需要手动重启该前台进程。Linux systemd 下服务命令使用 `systemctl --user`，需要可用的用户服务会话。
- Web 用户可在 **设置 -> 关于** 使用更新中心。CLI TUI 用户可输入 `/update`；确认更新后当前 TUI 会退出，以便 helper 安全替换已安装文件。

卸载行为：

- 可通过 `curl -fsSL https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.sh | sh` 或 `irm https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.ps1 | iex` 远程执行。
- 尽可能停止并卸载用户服务。
- 删除安装目录和 `slimebot` / `slimebot-cli` 命令入口。
- 删除 `~/.slimebot` 前会询问确认。
- `--yes` / `-Yes` 跳过确认并保留用户数据。
- `--purge` / `-Purge` 不询问，直接删除用户数据。
- `--keep-data` / `-KeepData` 显式保留用户数据。

## 测试

```bash
make test
# 或
go test ./...
```

前端和 CLI 测试：

```bash
npm --prefix frontend test
npm --prefix cli test
```

## Docker

```bash
make docker-build
make docker-run
```

## Docker Compose

```bash
make compose-up
make compose-down
```

## CLI 内置命令

- `/new` 新建会话
- `/session` 切换或删除会话
- `/model` 设置默认模型
- `/skills` 查看或移除 Skills
- `/mcp` MCP 增删改查，支持多行编辑
- `/approval` 切换审批模式
- `/effort` 设置思考等级
- `/plan` 切换规划模式
- `/update` 检查并应用 Release 更新
- `/help` 帮助

## 工具沙盒

SlimeBot 会对命令执行、文件工具和内置 HTTP 请求使用同一套沙盒策略。

- `read-only`：允许读取文件，禁止写入文件。
- `workspace-write`：默认模式。允许读取文件，并允许写入服务工作目录和额外配置的可写根目录。
- `danger-full-access`：保留不受限宿主机执行能力，应仅在明确需要时启用。
- 如果配置了 deny path，拒绝规则始终优先于可写根目录。
- 命令执行在 macOS 上通过 `/usr/bin/sandbox-exec` 执行，在 Linux 上通过 `bubblewrap`（`bwrap`）执行。Windows 暂未实现 OS 级沙盒。
- 网络访问由沙盒网络开关控制。
- 工具调用主动请求提升沙盒权限时会被视为单次权限提升请求。

## 数据目录

```text
~/.slimebot/
  config.cfg
  AGENTS.md
  skills/
  storage/
    data.db
    chat_uploads/
    update-status.json
```

- `config.cfg`：运行时配置文件
- `AGENTS.md`：全局 Agent 指令文件
- `storage/data.db`：SQLite 主数据库
- `storage/chat_uploads`：聊天附件
- `storage/update-status.json`：最近一次更新 helper 状态
- `skills/`：已安装 Skills

## 配置文件

SlimeBot 各组件会读取下列变量：

- `SERVER_PORT`：HTTP 端口，默认 `6247`
- `FRONTEND_PORT`：Vite 开发服务端口，默认 `7391`
- `DB_PATH`：SQLite 路径，默认 `~/.slimebot/storage/data.db`
- `SKILLS_ROOT`：Skills 根目录，默认 `~/.slimebot/skills`
- `CHAT_UPLOAD_ROOT`：附件目录，默认 `~/.slimebot/storage/chat_uploads`
- `CONTEXT_HISTORY_ROUNDS`：历史轮数配置，默认 `20`
- `DEFAULT_CONTEXT_SIZE`：新模型配置默认上下文大小，默认 `1000000`
- `FRONTEND_ORIGIN`：与 Vite 联调时设为 `http://localhost:7391`；生产同源可留空
- `WEB_SEARCH_API_KEY`：Tavily API Key
- `JWT_SECRET`：Web 服务模式必填
- `JWT_EXPIRE`：JWT 过期时间，单位分钟，默认 `21600`
- `SLIMEBOT_REPO`：安装/更新时覆盖 Release 仓库，默认 `natsuz0ra/SlimeBot`

示例：

```env
SERVER_PORT=6247
FRONTEND_PORT=7391
DB_PATH=~/.slimebot/storage/data.db
SKILLS_ROOT=~/.slimebot/skills
CHAT_UPLOAD_ROOT=~/.slimebot/storage/chat_uploads
WEB_SEARCH_API_KEY=YOUR_TAVILY_API_KEY
JWT_SECRET=CHANGE_ME_TO_A_RANDOM_SECRET
JWT_EXPIRE=21600
```

前端开发也可以使用 `frontend/.env`：

```env
VITE_API_BASE_URL=http://localhost:6247
VITE_WS_URL=ws://localhost:6247
```

## AGENTS.md 指令

- 全局指令位于 `~/.slimebot/AGENTS.md`。
- Web 设置页 AGENTS 标签可读取和保存全局指令。
- CLI 模式会额外读取项目级 `AGENTS.md`，并从 Git 仓库根目录到当前工作目录逐级合并。
- Server/Web 普通会话只注入全局 AGENTS 指令。

## 记忆机制

SlimeBot 现在有两层记忆：

- 会话上下文压缩是按会话存储在 SQLite 中的压缩摘要。如果完整历史超出所选模型配置的 `contextSize`，SlimeBot 会调用当前模型生成压缩摘要，写入 `session_context_summaries`，并在后续请求中以隐藏 `<context_summary>` 形式注入。
- 长期文件记忆会把代理长期笔记写入 `~/.slimebot/memories/MEMORY.md`，把用户画像写入 `~/.slimebot/memories/USER.md`。内置 `memory` 工具会在 Web/CLI 普通对话中静默更新这些文件，下一轮对话构建上下文时再注入最新快照。

长期记忆可在 Web 设置页的“记忆”标签管理，也可以在 CLI 中使用 `/memory` 和 `/memory reset memory|user|all` 查看或重置。
