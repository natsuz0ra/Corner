#!/usr/bin/env sh
set -eu

INSTALL_DIR="${SLIMEBOT_INSTALL_DIR:-"$HOME/.local/share/slimebot"}"
SHIM_DIR="${SLIMEBOT_BIN_DIR:-"$HOME/.local/bin"}"
DATA_DIR="${SLIMEBOT_HOME:-"$HOME/.slimebot"}"

YES=0
PURGE=0
KEEP_DATA=0

usage() {
  cat <<'EOF'
Usage: ./uninstall.sh [--yes] [--purge] [--keep-data] [--help]
       curl -fsSL https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.sh | sh

Uninstalls SlimeBot service, installed files, and command shims.

Options:
  --yes        Run non-interactively and keep user data.
  --purge      Run non-interactively and delete user data (~/.slimebot by default).
  --keep-data  Keep user data without prompting.
  --help       Show this help.

Environment:
  SLIMEBOT_INSTALL_DIR  Install directory (default: ~/.local/share/slimebot)
  SLIMEBOT_BIN_DIR      Command shim directory (default: ~/.local/bin)
  SLIMEBOT_HOME         User data directory (default: ~/.slimebot)

Remote examples:
  curl -fsSL https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.sh | sh
  curl -fsSL https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.sh | sh -s -- --yes
  curl -fsSL https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.sh | sh -s -- --purge
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --yes)
      YES=1
      ;;
    --purge)
      PURGE=1
      ;;
    --keep-data)
      KEEP_DATA=1
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

if [ "$PURGE" -eq 1 ] && [ "$KEEP_DATA" -eq 1 ]; then
  echo "--purge and --keep-data cannot be used together" >&2
  exit 2
fi

slimebot_bin=""
if [ -x "$INSTALL_DIR/bin/slimebot" ]; then
  slimebot_bin="$INSTALL_DIR/bin/slimebot"
elif command -v slimebot >/dev/null 2>&1; then
  slimebot_bin="$(command -v slimebot)"
fi

run_service_action() {
  action="$1"
  if [ -z "$slimebot_bin" ]; then
    echo "Skipping service $action: slimebot command not found"
    return 0
  fi
  if "$slimebot_bin" service "$action"; then
    echo "Service $action completed"
    return 0
  fi
  echo "Service $action failed or was not needed; continuing"
  return 0
}

run_service_action stop
run_service_action uninstall

rm -f "$SHIM_DIR/slimebot" "$SHIM_DIR/slimebot-cli"
rm -rf "$INSTALL_DIR"

delete_data=0
if [ "$PURGE" -eq 1 ]; then
  delete_data=1
elif [ "$KEEP_DATA" -eq 1 ] || [ "$YES" -eq 1 ]; then
  delete_data=0
elif [ -d "$DATA_DIR" ]; then
  printf "Delete SlimeBot user data at %s? This removes config, database, uploads, and skills. [y/N] " "$DATA_DIR"
  read answer || answer=""
  case "$answer" in
    y|Y|yes|YES|Yes)
      delete_data=1
      ;;
    *)
      delete_data=0
      ;;
  esac
fi

if [ "$delete_data" -eq 1 ]; then
  rm -rf "$DATA_DIR"
  echo "Deleted user data: $DATA_DIR"
else
  echo "Kept user data: $DATA_DIR"
fi

echo "Removed install directory: $INSTALL_DIR"
echo "Removed command shims from: $SHIM_DIR"
echo "SlimeBot uninstall complete"
