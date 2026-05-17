# Manual Deployment

This page covers source-based development, production builds, Docker, configuration, sandbox behavior, and runtime data layout. For the recommended end-user install path, see the main [README](../README.md).

## Architecture

- **Production:** one Go binary serves REST/WebSocket and embeds the web UI from `web/dist`.
- **Development:** `npm run dev` runs the Go server and Vite. Vite proxies `/api` and `/ws` to the backend port, `6247` by default.
- **Data:** SQLite defaults to `~/.slimebot/storage/data.db`.
- **Stack:** Go backend, Vue 3 web app, React + Ink CLI.

## Development Startup

Default ports: backend **6247**, Vite **7391**.

```bash
make deps
npm run dev
```

Manual dependency install:

```bash
npm install
npm install --prefix frontend
npm run dev
```

On first run, `~/.slimebot/config.cfg` is created if missing. If an older `~/.slimebot/.env` exists and `config.cfg` does not, SlimeBot copies it to `config.cfg` and keeps the old file untouched.

## Build From Source

Production build:

```bash
npm run build
# or
make build
```

Run the built Web service:

```bash
./slimebot server
```

Run the CLI TUI from source:

```bash
npm run cli
```

Create local Release archives:

```bash
make package
```

Release archives include `install.sh` / `install.ps1` and `uninstall.sh` / `uninstall.ps1`. The install scripts honor `SLIMEBOT_INSTALL_DIR` and `SLIMEBOT_BIN_DIR`; the uninstall scripts use the same variables plus `SLIMEBOT_HOME` for the user data directory.

The packaging script also writes standalone `dist/install.sh`, `dist/install.ps1`, `dist/uninstall.sh`, and `dist/uninstall.ps1` assets. When the install scripts are run outside an extracted Release archive, they resolve the latest GitHub Release, download the matching platform archive, and then run the installer inside that archive. Set `SLIMEBOT_VERSION=v1.26.1` to install a specific release tag, or `SLIMEBOT_REPO=owner/repo` for forks. The uninstall scripts can be run remotely without downloading a Release archive; they remove the local install using the configured install paths.

Update behavior:

- `slimebot update --check` checks the latest stable GitHub Release. The repository source follows the installer: `SLIMEBOT_REPO` first, otherwise `natsuz0ra/SlimeBot`.
- `slimebot update --yes` starts a detached helper process that downloads the matching Release archive, extracts it, runs the packaged installer, and writes progress to `~/.slimebot/storage/update-status.json`.
- `slimebot update --version vX.Y.Z --yes` installs a specific Release tag. This is also the recovery command shown when the current build is `dev`, empty, or not parseable as a version.
- In Web service mode, the helper attempts `slimebot service stop`, installs the update, then attempts `slimebot service start`. If SlimeBot is running as a foreground `slimebot server` process without a service, restart that process manually after the helper finishes.
- Web users can use the update center in **Settings -> About**. CLI TUI users can run `/update`; applying an update exits the current TUI so the helper can replace the installed files safely.

Uninstall behavior:

- Can be run remotely with `curl -fsSL https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.sh | sh` or `irm https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.ps1 | iex`.
- Stops and uninstalls the system service when possible.
- Removes the install directory and `slimebot` / `slimebot-cli` command shims.
- Prompts before deleting `~/.slimebot`.
- `--yes` / `-Yes` keeps user data and skips prompts.
- `--purge` / `-Purge` deletes user data without prompting.
- `--keep-data` / `-KeepData` explicitly preserves user data.

## Tests

```bash
make test
# or
go test ./...
```

Frontend and CLI tests:

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

## CLI Commands

- `/new` — new session
- `/session` — switch or delete sessions
- `/model` — set default model
- `/skills` — view or remove skills
- `/mcp` — MCP CRUD with multiline editor
- `/approval` — toggle approval mode
- `/effort` — set thinking level
- `/plan` — toggle plan mode
- `/update` — check and apply Release updates
- `/help` — help

## Tool Sandbox

SlimeBot applies one sandbox policy across command execution, file tools, and built-in HTTP requests.

- `read-only` allows file reads but blocks file writes.
- `workspace-write` is the default. It allows reads and writes under the server working directory plus configured writable roots.
- `danger-full-access` keeps the legacy unrestricted host behavior and should only be enabled intentionally.
- Denied paths, when configured, take priority over writable roots.
- Command execution runs through `/usr/bin/sandbox-exec` on macOS and `bubblewrap` (`bwrap`) on Linux. Windows OS-level sandboxing is not implemented yet.
- Network access is controlled by the sandbox network setting.
- Tool calls that request elevated sandbox permissions are treated as one-call escalation requests.

## Data Layout

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

- `config.cfg` — runtime configuration
- `AGENTS.md` — global Agent instructions
- `storage/data.db` — SQLite database
- `storage/chat_uploads` — chat attachments
- `storage/update-status.json` — latest update helper status
- `skills/` — installed skills

## Configuration

Variables read by SlimeBot components:

- `SERVER_PORT` — HTTP port, default `6247`
- `FRONTEND_PORT` — Vite development server port, default `7391`
- `DB_PATH` — SQLite path, default `~/.slimebot/storage/data.db`
- `SKILLS_ROOT` — skills root, default `~/.slimebot/skills`
- `CHAT_UPLOAD_ROOT` — uploads, default `~/.slimebot/storage/chat_uploads`
- `CONTEXT_HISTORY_ROUNDS` — retained history-round setting, default `20`
- `DEFAULT_CONTEXT_SIZE` — default context size for new model configs, default `1000000`
- `FRONTEND_ORIGIN` — set to `http://localhost:7391` when using Vite; empty for same-origin production
- `WEB_SEARCH_API_KEY` — Tavily API key for web search
- `JWT_SECRET` — required in Web server mode
- `JWT_EXPIRE` — JWT lifetime in minutes, default `21600`
- `SLIMEBOT_REPO` — Release repository for installer/updater overrides, default `natsuz0ra/SlimeBot`

Example:

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

Frontend development can also use `frontend/.env`:

```env
VITE_API_BASE_URL=http://localhost:6247
VITE_WS_URL=ws://localhost:6247
```

## AGENTS.md Instructions

- Global instructions live at `~/.slimebot/AGENTS.md`.
- The Web settings AGENTS tab reads and saves global instructions.
- CLI mode also reads project-level `AGENTS.md` files and merges them from the Git repository root to the current working directory.
- Server/Web sessions inject only global AGENTS instructions.

## Memory Model

Memory is a SQLite-backed compact summary for each chat session. If the full history fits under the selected model config's `contextSize`, SlimeBot sends the full history. If it exceeds the window, SlimeBot asks the current model to generate a compact summary, stores it in `session_context_summaries`, and injects it later as a hidden `<context_summary>`.
