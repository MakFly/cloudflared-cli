#!/usr/bin/env bash
set -euo pipefail

# ─── Constants ───────────────────────────────────────────────────────────────
REPO="kev/cloudflared-cli"
BINARY="cloudflared-project"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASES_URL="https://github.com/${REPO}/releases/download"

# ─── Colors ──────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  CYAN='\033[0;36m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' CYAN='' BOLD='' RESET=''
fi

ok()   { printf "${GREEN}  %s${RESET} %s\n" "✓" "$*"; }
fail() { printf "${RED}  %s${RESET} %s\n"  "✗" "$*" >&2; }
info() { printf "${CYAN}  →${RESET} %s\n"  "$*"; }
step() { printf "\n${BOLD}%s${RESET}\n" "$*"; }

die() {
  fail "$*"
  exit 1
}

# ─── Helpers ─────────────────────────────────────────────────────────────────
need() {
  command -v "$1" >/dev/null 2>&1 || die "Required command not found: $1"
}

detect_os() {
  local raw
  raw="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$raw" in
    darwin) echo "darwin" ;;
    linux)  echo "linux"  ;;
    *)      die "Unsupported OS: $raw" ;;
  esac
}

detect_arch() {
  local raw
  raw="$(uname -m)"
  case "$raw" in
    x86_64)           echo "amd64" ;;
    amd64)            echo "amd64" ;;
    aarch64|arm64)    echo "arm64" ;;
    *)                die "Unsupported architecture: $raw" ;;
  esac
}

install_dir() {
  if [ "$(id -u)" -eq 0 ]; then
    echo "/usr/local/bin"
  else
    echo "${HOME}/.local/bin"
  fi
}

fetch_latest_tag() {
  local tag
  tag="$(curl -fsSL "${API_URL}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  [ -n "$tag" ] || die "Could not fetch latest release tag from ${API_URL}"
  echo "$tag"
}

# ─── Install ─────────────────────────────────────────────────────────────────
do_install() {
  need curl
  need tar
  need uname

  step "Installing ${BINARY}"

  local os arch tag tarball url dest tmpdir

  os="$(detect_os)"
  arch="$(detect_arch)"
  info "Detected platform: ${os}/${arch}"

  tag="$(fetch_latest_tag)"
  info "Latest release: ${tag}"

  tarball="${BINARY}-${os}-${arch}.tar.gz"
  url="${RELEASES_URL}/${tag}/${tarball}"
  info "Downloading: ${url}"

  dest="$(install_dir)"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  if ! curl -fsSL "${url}" -o "${tmpdir}/${tarball}"; then
    die "Download failed: ${url}"
  fi
  ok "Download complete"

  tar -xzf "${tmpdir}/${tarball}" -C "${tmpdir}"

  # The tarball may contain the binary directly or inside a directory; find it.
  local bin_path
  bin_path="$(find "${tmpdir}" -type f -name "${BINARY}" | head -1)"
  [ -n "$bin_path" ] || die "Binary '${BINARY}' not found inside tarball"

  chmod +x "${bin_path}"

  # Create destination dir if needed (e.g. ~/.local/bin)
  if [ ! -d "${dest}" ]; then
    mkdir -p "${dest}"
    info "Created directory: ${dest}"
  fi

  # Use sudo only when installing to a system path and not root
  if [ "$(id -u)" -ne 0 ] && [[ "${dest}" == /usr/* ]]; then
    sudo mv "${bin_path}" "${dest}/${BINARY}"
  else
    mv "${bin_path}" "${dest}/${BINARY}"
  fi

  ok "Installed to ${dest}/${BINARY}"

  # PATH hint when dest is not in PATH
  if ! echo ":${PATH}:" | grep -q ":${dest}:"; then
    printf "\n${YELLOW}  ! ${dest} is not in your PATH.${RESET}\n"
    printf "    Add to your shell profile:\n"
    printf "      ${BOLD}export PATH=\"${dest}:\$PATH\"${RESET}\n\n"
  fi

  step "Verifying installation"
  if "${dest}/${BINARY}" version >/dev/null 2>&1; then
    ok "$("${dest}/${BINARY}" version)"
  else
    die "Verification failed — '${BINARY} version' exited with an error"
  fi

  printf "\n${GREEN}${BOLD}Installation complete.${RESET}\n\n"
}

# ─── Uninstall ────────────────────────────────────────────────────────────────
do_uninstall() {
  step "Uninstalling ${BINARY}"

  local dest removed=0

  for dir in "${HOME}/.local/bin" "/usr/local/bin"; do
    local target="${dir}/${BINARY}"
    if [ -f "${target}" ]; then
      info "Found: ${target}"
      if [ "$(id -u)" -ne 0 ] && [[ "${dir}" == /usr/* ]]; then
        sudo rm -f "${target}"
      else
        rm -f "${target}"
      fi
      ok "Removed ${target}"
      removed=1
    fi
  done

  if [ "${removed}" -eq 0 ]; then
    fail "${BINARY} not found in ~/.local/bin or /usr/local/bin"
    exit 1
  fi

  printf "\n${GREEN}${BOLD}Uninstall complete.${RESET}\n\n"
}

# ─── Entry point ─────────────────────────────────────────────────────────────
main() {
  case "${1:-}" in
    --uninstall|-u) do_uninstall ;;
    --help|-h)
      printf "Usage: install.sh [--uninstall]\n"
      printf "\n"
      printf "  (no flag)     Install ${BINARY}\n"
      printf "  --uninstall   Remove ${BINARY}\n"
      ;;
    "")  do_install ;;
    *)   die "Unknown flag: $1. Use --help for usage." ;;
  esac
}

main "$@"
