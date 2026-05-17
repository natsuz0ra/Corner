#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git -C "${ROOT_DIR}" rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

TARGETS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

cd "${ROOT_DIR}"

npm --prefix frontend run build
npm --prefix cli run build

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"
cp scripts/install.sh scripts/install.ps1 scripts/uninstall.sh scripts/uninstall.ps1 "${DIST_DIR}/"
chmod +x "${DIST_DIR}/install.sh" "${DIST_DIR}/uninstall.sh"

for target in "${TARGETS[@]}"; do
  goos="${target%%/*}"
  goarch="${target##*/}"
  name="slimebot-${VERSION}-${goos}-${goarch}"
  package_dir="${DIST_DIR}/${name}"
  bin_dir="${package_dir}/bin"
  exe_suffix=""
  archive="${DIST_DIR}/${name}.tar.gz"
  if [[ "${goos}" == "windows" ]]; then
    exe_suffix=".exe"
    archive="${DIST_DIR}/${name}.zip"
  fi

  mkdir -p "${bin_dir}" "${package_dir}/cli"

  env GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X slimebot/internal/version.Version=${VERSION} -X slimebot/internal/version.Commit=${COMMIT} -X slimebot/internal/version.Date=${BUILD_DATE}" \
    -o "${bin_dir}/slimebot${exe_suffix}" ./cmd/server

  cp cli/cli.cjs "${package_dir}/cli/cli.cjs"
  cp scripts/install.sh "${package_dir}/install.sh"
  cp scripts/install.ps1 "${package_dir}/install.ps1"
  cp scripts/uninstall.sh "${package_dir}/uninstall.sh"
  cp scripts/uninstall.ps1 "${package_dir}/uninstall.ps1"
  cp README.md README.zh-CN.md LICENSE "${package_dir}/"
  if [[ -d docs ]]; then
    cp -R docs "${package_dir}/docs"
  fi

  if [[ "${goos}" == "windows" ]]; then
    cat > "${bin_dir}/slimebot-cli.cmd" <<'EOF'
@echo off
"%~dp0slimebot.exe" cli %*
EOF
    (cd "${DIST_DIR}" && zip -qr "${archive}" "${name}")
  else
    cat > "${bin_dir}/slimebot-cli" <<'EOF'
#!/usr/bin/env sh
exec "$(dirname "$0")/slimebot" cli "$@"
EOF
    chmod +x "${bin_dir}/slimebot" "${bin_dir}/slimebot-cli" "${package_dir}/install.sh" "${package_dir}/uninstall.sh"
    (cd "${DIST_DIR}" && tar -czf "${archive}" "${name}")
  fi
done

echo "Release packages written to ${DIST_DIR}"
