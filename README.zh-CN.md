<p align="center">
  <img src="assets/title.png" alt="SlimeBot Logo" width="420" />
  <br /><br />
  <a href="README.md">English</a> | <strong>简体中文</strong>
</p>

# SlimeBot

个人练手的 Agent Demo，目标是搭建可扩展的 AI 会话应用雏形。采用 **Go** 后端、**Vue 3** Web 前端，以及 **React + Ink** 终端 CLI。

## 当前支持功能

- **会话与消息**
  - 会话列表、创建、重命名、删除
  - 按会话拉取历史消息
  - 基于 WebSocket 的实时流式回复
  - 会话标题自动生成与更新推送
  - 多模态能力
- **工具与 Agent**
  - Agent 多轮 tool call 执行链路
  - 审批模式支持：**标准模式**（敏感工具需确认）、**自动审查**（先由模型审查不确定的敏感工具）与**自动执行**（直接执行）
  - 命令执行、文件修改等敏感内置操作需用户确认，支持 Web、CLI、Telegram 等流程
  - 沙盒策略支持，覆盖命令执行、文件访问与内置 HTTP 请求，支持 `read-only`、`workspace-write`、`danger-full-access`
  - 工具结果写入会话历史并支持详情查看
  - 内置能力：命令行、网络请求、基于 Tavily 的网络搜索、待办事项
  - 支持面向代码编辑场景的文件读写能力，用于文本文件编辑场景
  - **子代理：**主 Agent 可将独立子任务交给内层 Agent，内层使用**隔离上下文**（不携带父会话聊天记录）。仅支持**一层嵌套**。子代理内的工具调用在 Web 与 CLI 中**嵌套展示**在父工具之下；历史记录会持久化父子关系，刷新会话后层级仍可还原。
- **规划与思考控制**
  - 规划模式（Plan Mode）：先产出计划，再审批后执行
  - 计划生命周期：生成、同意/拒绝、修改并重生成、审批后执行
  - 思考等级控制（`off` / `low` / `medium` / `high`）
  - Web 与 CLI 均支持思考流式事件展示与时间线呈现
- **记忆能力**
  - 按会话与模型配置保存隐藏压缩摘要
  - 上下文超出模型 `contextSize` 时自动压缩历史
  - Web 与 CLI 展示上下文用量和压缩状态
- **配置与扩展**
  - MCP 配置管理
  - Skills 上传安装、列表、删除与运行时激活
  - AGENTS.md 指令支持：全局指令可在 Web 设置页编辑；CLI 会额外读取项目级 `AGENTS.md`，并从 Git 仓库根目录到当前工作目录逐级合并后注入模型上下文
- **消息平台**（当前支持 Telegram）
  - 消息平台配置管理
  - 平台消息接入与回复
- **CLI TUI**
  - 独立 CLI（无头 Go 子进程 + Ink 终端界面），支持对话与基本配置

## UI 预览

### 登录页

![登录页预览](assets/login.png)

### 主页

![主页预览](assets/home.png)

### 会话页

![会话页预览](assets/chat.png)

### 规划模式

![规划模式](assets/plan.png)

### 工具执行

![工具执行](assets/tool_exec.png)

### 消息平台（Telegram）

<img src="assets/tg_chat.png" alt="消息平台预览" width="220" />

### CLI

<img src="assets/cli.png" alt="CLI" width="800" />

## 架构与技术栈

- **生产**：Go 进程同时提供 REST/WebSocket，并通过 `go:embed` 嵌入 `web/dist` 静态资源，单一可执行文件交付。
- **开发**：`npm run dev` 同时启动 Go 与 Vite；Vite 将 `/api`、`/ws` 代理到配置的后端端口（默认 `6247`）。
- **数据**：默认 SQLite，路径 `~/.slimebot/storage/data.db`，会话压缩摘要也持久化在其中。
- **记忆**：当前为会话内上下文压缩记忆。系统在需要时生成隐藏 `<context_summary>`，并与最新消息一起注入模型上下文。

**技术栈（概览）：** Go 后端 · Vue 3 Web 前端 · React + Ink CLI。

## 如何启动

默认端口：后端 **6247**，Vite **7391**。

在仓库根目录：

```bash
make deps
npm run dev
```

或手动安装依赖：

```bash
npm install
npm install --prefix frontend
npm run dev
```

首次启动会在缺失时创建 `~/.slimebot/config.cfg`；后续若嵌入式模板新增键名，会按需追加到现有文件。若存在旧的 `~/.slimebot/.env` 且不存在 `config.cfg`，SlimeBot 会复制旧配置到 `config.cfg`，并保留旧文件不动。

**首次登录（Web 服务模式）：** 若数据库中尚无用户，会种子一个默认账号（用户名 **`admin`**，密码 **`admin`**），并引导修改密码。除本机尝鲜外请尽快修改。

**生产构建**（生成嵌入前端的 `slimebot` 可执行文件）：

```bash
npm run build
# 或
make build
```

**仅运行后端**（需先完成前端构建以提供静态页）：

```bash
go run ./cmd/server/main.go
```

**CLI TUI：**

```bash
npm run cli
```

`make cli` 会安装 CLI 的 npm 依赖、构建 CLI（React + Ink）并生成 `slimebot-cli` 可执行文件（见 [Makefile](Makefile)）。

**测试：**

```bash
make test
# 或
go test ./...
```

**Docker：**

```bash
make docker-build
make docker-run
```

**Docker Compose：**

```bash
make compose-up
make compose-down
```

### CLI 内置命令

- `/new` 新建会话（懒创建，首次发送消息才真正建会话）
- `/session` 会话菜单（切换 / 删除）
- `/model` 模型菜单（切换全局默认模型）
- `/skills` 技能菜单（查看信息 / 删除）
- `/mcp` MCP 菜单（增删改查，内置多行编辑）
- `/approval` 切换审批模式（`standard` / `auto_review` / `auto`）
- `/effort` 设置思考等级（`off` / `low` / `medium` / `high`）
- `/plan` 切换规划模式（`on` / `off`）
- `/help` 帮助

## 工具沙盒

SlimeBot 会对命令执行、文件工具和内置 HTTP 请求使用同一套沙盒策略。

- `read-only`：允许读取文件，禁止写入文件。
- `workspace-write`：默认模式。允许读取文件，并允许写入服务工作目录和额外配置的可写根目录。
- `danger-full-access`：保留旧版不受限的宿主机执行能力，应仅在明确需要时启用。
- 如果配置了 deny path，拒绝规则始终优先于可写根目录。
- 命令执行在 macOS 上通过 `/usr/bin/sandbox-exec` 执行，在 Linux 上通过 `bubblewrap`（`bwrap`）执行。平台沙盒不可用时会失败关闭，不会静默退回宿主机裸执行。Windows 暂未实现 OS 级沙盒。
- 网络访问由沙盒网络开关控制。内置 HTTP 请求额外支持域名 allowlist；沙盒子进程首版支持网络开启/关闭。
- 工具调用主动请求提升沙盒权限时会被视为权限提升请求，审批结果只作用于当前这一次工具调用。

## 数据与资源目录（默认）

所有运行时数据默认集中在 `~/.slimebot`：

```text
~/.slimebot/
  config.cfg
  AGENTS.md
  skills/
  storage/
    data.db
    chat_uploads/
```

- `config.cfg`：运行时配置文件
- `AGENTS.md`：全局 Agent 指令文件，可通过 Web 设置页的 AGENTS 标签读取和保存
- `storage/data.db`：SQLite 主数据库
- `storage/chat_uploads`：聊天附件
- `skills`：Skills 存储目录

## AGENTS.md 指令

- 全局指令文件位于 `~/.slimebot/AGENTS.md`，用于为所有会话提供长期工作规则。
- Web 设置页的 AGENTS 标签会通过 `GET /api/agents-instructions` 读取全局指令，并通过 `PUT /api/agents-instructions` 保存更新。
- CLI 模式会额外读取当前项目中的 `AGENTS.md`。如果当前工作目录位于 Git 仓库内，会按“仓库根目录 -> 当前工作目录”的顺序合并沿途非空的 `AGENTS.md` 文件。
- Server/Web 普通会话只注入全局 AGENTS 指令，不读取项目局部文件。
- AGENTS 内容会进入稳定系统提示词；当全局或项目指令变化时，系统提示词缓存会随之刷新。

## 记忆存储（工作机制）

- 记忆不是独立的 Markdown 文件或全文索引，而是存储在 SQLite 中的会话压缩摘要。
- 当完整历史、系统提示词、运行环境信息和工具回放估算后仍低于模型配置的 `contextSize` 时，会直接发送完整历史。
- 当上下文超过 `contextSize` 时，系统调用当前模型生成压缩摘要，写入 `session_context_summaries`，并在后续请求中以隐藏 `<context_summary>` 形式注入。
- 摘要按 `sessionId + modelConfigId` 区分；同一会话切换不同模型配置时会使用对应配置的摘要。
- 已有摘要可复用；如果摘要后的新消息再次超窗，会把旧摘要与新增消息滚动压缩成新的摘要。
- 压缩摘要仅作为同一会话的连续性上下文使用；如果与新的用户输入冲突，系统提示要求优先遵循新消息。
- 最新一条用户输入会被保护：若它单独就超过上下文窗口，会直接报错，提示缩短输入或调大上下文大小。
- Web 与 CLI 会通过 `context_usage` / `context_compacted` 事件展示已用 token、可用比例和是否已压缩。

## 配置文件（`~/.slimebot/config.cfg`）

SlimeBot 各组件会读取下列变量（括号内为默认值或说明）：

- `SERVER_PORT`：服务端口，默认 `6247`
- `FRONTEND_PORT`：Vite 开发服务端口，默认 `7391`
- `DB_PATH`：SQLite 文件路径，默认 `~/.slimebot/storage/data.db`
- `SKILLS_ROOT`：Skills 根目录，默认 `~/.slimebot/skills`
- `CHAT_UPLOAD_ROOT`：聊天附件目录，默认 `~/.slimebot/storage/chat_uploads`
- `CONTEXT_HISTORY_ROUNDS`：历史轮数配置保留项，默认 `20`，内部限制为 `5` 到 `50`
- `DEFAULT_CONTEXT_SIZE`：新建模型配置的默认上下文大小，默认 `1000000`
- `FRONTEND_ORIGIN`：与 Vite 联调时设为 `http://localhost:7391`；生产同源可留空
- `WEB_SEARCH_API_KEY`：Tavily API Key，供网络搜索使用
- `JWT_SECRET`：**服务端模式必填**，未配置将启动失败（CLI 无头模式可自动生成）
- `JWT_EXPIRE`：JWT 过期时间（单位：分钟，默认 `21600` 即约 15 天）

首次启动生成的 `config.cfg` 与嵌入式模板一致，见 [internal/runtime/env.template](internal/runtime/env.template)。其他键可按需自行追加。旧的 `~/.slimebot/.env` 会在首次启动时复制迁移到 `config.cfg`。

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

# CONTEXT_HISTORY_ROUNDS=20
# DEFAULT_CONTEXT_SIZE=1000000

# FRONTEND_ORIGIN=http://localhost:7391
```

### 前端配置：`frontend/.env`

- `VITE_API_BASE_URL`：后端 HTTP 地址（例如 `http://localhost:6247`）
- `VITE_WS_URL`：后端 WebSocket 地址（例如 `ws://localhost:6247`）
- `FRONTEND_PORT`：Vite 开发服务端口；从进程环境变量或 `~/.slimebot/config.cfg` 读取

示例：

```env
VITE_API_BASE_URL=http://localhost:6247
VITE_WS_URL=ws://localhost:6247
```

## 功能状态与待办

### 已完成

- 会话管理与 WebSocket 流式回复（含错误、工具调用、子代理与思考事件）
- Agent 工具、审批与沙盒约束，覆盖命令执行、文件访问和内置 HTTP 请求
- 规划模式：计划生成、同意/拒绝/修改流程，以及审批后执行
- 思考等级控制（`off` / `low` / `medium` / `high`）与流式思考展示
- 子代理 / 嵌套 Agent、嵌套工具 UI，以及工具历史中的父子关联持久化
- MCP 与 Skills
- AGENTS.md 全局/项目指令支持，以及设置页 AGENTS 编辑能力
- 基于 SQLite 的会话压缩摘要、上下文用量统计与隐藏上下文注入
- Telegram 集成
- 多模态支持
- JWT 认证与默认管理员种子

### 待完成功能

- 更多消息平台接入（如 Discord、Slack 等）

## 许可

本项目以 [MIT 许可证](LICENSE) 授权。
