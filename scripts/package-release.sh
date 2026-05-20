#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git -C "${ROOT_DIR}" rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
export COPYFILE_DISABLE=1
export COPY_EXTENDED_ATTRIBUTES_DISABLE=1

TARGETS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

assert_cli_bundle_self_contained() {
  local bundle="$1"
  local deps=(
    "ink"
    "react"
    "react/jsx-runtime"
    "ink-text-input"
  )

  for dep in "${deps[@]}"; do
    if grep -Fq "from \"$dep\"" "$bundle" || grep -Fq "from '$dep'" "$bundle"; then
      echo "CLI bundle still imports runtime dependency: $dep" >&2
      echo "Update cli/tsup.config.ts so release packages do not require cli/node_modules." >&2
      exit 1
    fi
  done
}

ripgrep_binary_name() {
  local goos="$1"
  if [[ "${goos}" == "windows" ]]; then
    echo "rg.exe"
  else
    echo "rg"
  fi
}

copy_ripgrep_vendor() {
  local goos="$1"
  local goarch="$2"
  local bin_dir="$3"
  local platform="${goos}-${goarch}"
  local rg_name
  rg_name="$(ripgrep_binary_name "${goos}")"
  local source="${ROOT_DIR}/third_party/ripgrep/${platform}/${rg_name}"
  local dest_dir="${bin_dir}/vendor/ripgrep/${platform}"

  if [[ ! -f "${source}" ]]; then
    echo "Missing bundled ripgrep binary: ${source}" >&2
    echo "Place ${rg_name} at third_party/ripgrep/${platform}/ before packaging ${platform}." >&2
    exit 1
  fi

  mkdir -p "${dest_dir}"
  cp "${source}" "${dest_dir}/${rg_name}"
  if [[ "${goos}" != "windows" ]]; then
    chmod +x "${dest_dir}/${rg_name}"
  fi
}

cd "${ROOT_DIR}"

npm --prefix frontend run build
npm --prefix cli run build

if [[ ! -f cli/cli.cjs ]]; then
  echo "Missing CLI entry: cli/cli.cjs" >&2
  exit 1
fi
if [[ ! -f cli/dist/index.js ]]; then
  echo "Missing CLI bundle: cli/dist/index.js" >&2
  exit 1
fi
assert_cli_bundle_self_contained "cli/dist/index.js"

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
  copy_ripgrep_vendor "${goos}" "${goarch}" "${bin_dir}"

  cp cli/cli.cjs "${package_dir}/cli/cli.cjs"
  cp -R cli/dist "${package_dir}/cli/dist"
  cp scripts/install.sh "${package_dir}/install.sh"
  cp scripts/install.ps1 "${package_dir}/install.ps1"
  cp scripts/uninstall.sh "${package_dir}/uninstall.sh"
  cp scripts/uninstall.ps1 "${package_dir}/uninstall.ps1"
  cp README.md README.zh-CN.md LICENSE "${package_dir}/"
  if [[ -d docs ]]; then
    cp -R docs "${package_dir}/docs"
  fi
  find "${package_dir}" -name '._*' -delete

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
