#!/usr/bin/env sh
set -eu

SOURCE_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
INSTALL_DIR="${SLIMEBOT_INSTALL_DIR:-"$HOME/.local/share/slimebot"}"
SHIM_DIR="${SLIMEBOT_BIN_DIR:-"$HOME/.local/bin"}"

mkdir -p "$INSTALL_DIR" "$SHIM_DIR"

rm -rf "$INSTALL_DIR/bin" "$INSTALL_DIR/cli"
cp -R "$SOURCE_DIR/bin" "$SOURCE_DIR/cli" "$INSTALL_DIR/"
if [ -d "$SOURCE_DIR/docs" ]; then
  cp -R "$SOURCE_DIR/docs" "$INSTALL_DIR/"
fi
cp "$SOURCE_DIR/README.md" "$SOURCE_DIR/README.zh-CN.md" "$SOURCE_DIR/LICENSE" "$INSTALL_DIR/"

cat > "$SHIM_DIR/slimebot" <<EOF
#!/usr/bin/env sh
exec "$INSTALL_DIR/bin/slimebot" "\$@"
EOF

cat > "$SHIM_DIR/slimebot-cli" <<EOF
#!/usr/bin/env sh
exec "$INSTALL_DIR/bin/slimebot" cli "\$@"
EOF

chmod +x "$INSTALL_DIR/bin/slimebot" "$SHIM_DIR/slimebot" "$SHIM_DIR/slimebot-cli"

mkdir -p "$HOME/.slimebot"
if [ ! -f "$HOME/.slimebot/config.cfg" ]; then
  "$INSTALL_DIR/bin/slimebot" help >/dev/null
  cat > "$HOME/.slimebot/config.cfg" <<'EOF'
SERVER_PORT=6247
FRONTEND_PORT=7391
DB_PATH=~/.slimebot/storage/data.db
SKILLS_ROOT=~/.slimebot/skills
CHAT_UPLOAD_ROOT=~/.slimebot/storage/chat_uploads
WEB_SEARCH_API_KEY=YOUR_TAVILY_API_KEY
JWT_SECRET=CHANGE_ME_TO_A_RANDOM_SECRET
JWT_EXPIRE=21600
EOF
fi

echo "SlimeBot installed to $INSTALL_DIR"
echo "Command shims installed to $SHIM_DIR"
echo "Run: slimebot"
echo "If your shell cannot find slimebot, add this to PATH: $SHIM_DIR"
echo "Before starting the web service, set JWT_SECRET in $HOME/.slimebot/config.cfg"
