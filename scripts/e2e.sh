#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────────────────────
# e2e.sh — End-to-end test for cloudflared-project CLI.
#
# Usage:
#   ./scripts/e2e.sh            Install cloudflared, build CLI, run all tests
#   ./scripts/e2e.sh --remove   Remove everything: cloudflared, test data,
#                               built binary, projects, brew cache
# ─────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_ROOT/cloudflared-project"
TEST_PROJECT="e2e-testapp"
TEST_DOMAIN="testapp.example.com"
MARKER_FILE="$PROJECT_ROOT/.e2e-installed-cloudflared"

# Runtime state
QUICK_TUNNEL_PID=""
LOCAL_SERVER_PID=""
TUNNEL_OUTPUT=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
DIM='\033[2m'
BOLD='\033[1m'
RESET='\033[0m'

passed=0
failed=0
skipped=0

# ─────────────────────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────────────────────

log()     { echo -e "${CYAN}→${RESET} $*"; }
ok()      { echo -e "${GREEN}✓${RESET} $*"; ((passed++)); }
fail()    { echo -e "${RED}✗${RESET} $*"; ((failed++)); }
skip()    { echo -e "${YELLOW}⊘${RESET} $*"; ((skipped++)); }
section() { echo -e "\n${BOLD}${CYAN}━━━ $* ━━━${RESET}"; }
die()     { echo -e "${RED}$*${RESET}" >&2; exit 1; }

assert_exit_0() {
  local desc="$1"; shift
  if "$@" > /dev/null 2>&1; then ok "$desc"
  else fail "$desc"; fi
}

assert_exit_nonzero() {
  local desc="$1"; shift
  if "$@" > /dev/null 2>&1; then fail "$desc (expected failure)"
  else ok "$desc"; fi
}

assert_contains() {
  local desc="$1" expected="$2"; shift 2
  local output
  output=$("$@" 2>&1) || true
  if echo "$output" | grep -q "$expected"; then ok "$desc"
  else
    fail "$desc — expected '$expected'"
    echo -e "${DIM}  got: $(echo "$output" | head -3)${RESET}"
  fi
}

assert_file_exists() {
  if [[ -f "$2" ]]; then ok "$1"
  else fail "$1 — not found: $2"; fi
}

assert_file_contains() {
  if [[ -f "$2" ]] && grep -q "$3" "$2"; then ok "$1"
  else fail "$1 — '$3' not in $2"; fi
}

# ─────────────────────────────────────────────────────────────
# --remove : nuke everything
# ─────────────────────────────────────────────────────────────

do_remove() {
  section "REMOVE: Full cleanup"

  # 1. Kill any leftover tunnel / server from a crashed run
  log "Stopping stale processes..."
  pkill -f "cloudflared tunnel --url" 2>/dev/null || true
  pkill -f "python3.*18923" 2>/dev/null || true

  # Clean up stale e2e tunnel (only if cloudflared is still installed)
  if command -v cloudflared &>/dev/null; then
    cloudflared tunnel delete "$TEST_PROJECT" 2>/dev/null || true
  fi

  # 2. Remove test project data
  local proj_dir="$HOME/.cloudflared/projects/$TEST_PROJECT"
  if [[ -d "$proj_dir" ]]; then
    log "Removing test project: $proj_dir"
    rm -rf "$proj_dir"
  fi

  # Remove projects dir if empty
  if [[ -d "$HOME/.cloudflared/projects" ]]; then
    rmdir "$HOME/.cloudflared/projects" 2>/dev/null || true
  fi

  # Remove local test project
  rm -rf "$PROJECT_ROOT/.cloudflared-project" 2>/dev/null || true

  # 3. Remove built binary
  if [[ -f "$BINARY" ]]; then
    log "Removing binary: $BINARY"
    rm -f "$BINARY"
  fi

  # 4. Uninstall cloudflared (only if we installed it)
  if [[ -f "$MARKER_FILE" ]]; then
    log "Uninstalling cloudflared (installed by e2e.sh)..."
    brew uninstall --force cloudflared 2>/dev/null || true
    brew autoremove 2>/dev/null || true
    brew cleanup cloudflared 2>/dev/null || true
    rm -f "$MARKER_FILE"

    # Remove cert.pem since we installed cloudflared ourselves
    if [[ -f "$HOME/.cloudflared/cert.pem" ]]; then
      log "Removing cert.pem (cloudflared was installed by e2e.sh)"
      rm -f "$HOME/.cloudflared/cert.pem"
    fi

    # Verify
    if command -v cloudflared &>/dev/null; then
      echo -e "${YELLOW}! cloudflared still in PATH after uninstall — check manually${RESET}"
    else
      log "cloudflared uninstalled"
    fi
  else
    # Don't remove cert.pem if cloudflared was pre-existing — user may need it
    log "cloudflared was not installed by e2e.sh — skipping uninstall"
  fi

  # 5. Clean up cloudflared runtime data it may have created
  #    (only the stuff cloudflared auto-creates, not user creds)
  rm -rf "$HOME/.cloudflared/projects" 2>/dev/null || true
  # Don't touch ~/.cloudflared/*.json or cert.pem — those are user credentials

  # 6. Remove temp files
  rm -f /tmp/e2e-tunnel-*.log 2>/dev/null || true

  echo ""
  echo -e "${GREEN}Cleanup complete.${RESET}"
  exit 0
}

# ─────────────────────────────────────────────────────────────
# Trap: clean processes on unexpected exit during tests
# ─────────────────────────────────────────────────────────────

trap_cleanup() {
  if [[ -n "$QUICK_TUNNEL_PID" ]]; then
    kill "$QUICK_TUNNEL_PID" 2>/dev/null || true
    wait "$QUICK_TUNNEL_PID" 2>/dev/null || true
  fi
  if [[ -n "$LOCAL_SERVER_PID" ]]; then
    kill "$LOCAL_SERVER_PID" 2>/dev/null || true
    wait "$LOCAL_SERVER_PID" 2>/dev/null || true
  fi
  rm -f "$TUNNEL_OUTPUT" 2>/dev/null || true
}

print_results() {
  section "RESULTS"
  echo -e "  ${GREEN}Passed:  $passed${RESET}"
  [[ $failed  -gt 0 ]] && echo -e "  ${RED}Failed:  $failed${RESET}" || echo -e "  Failed:  0"
  [[ $skipped -gt 0 ]] && echo -e "  ${YELLOW}Skipped: $skipped${RESET}" || echo -e "  Skipped: 0"
  echo ""
  if [[ $failed -gt 0 ]]; then
    echo -e "${RED}Some tests failed.${RESET}"
    echo -e "${DIM}Run './scripts/e2e.sh --remove' to clean up.${RESET}"
    exit 1
  else
    echo -e "${GREEN}All tests passed.${RESET}"
    echo -e "${DIM}Run './scripts/e2e.sh --remove' to uninstall cloudflared and clean up.${RESET}"
  fi
}

# ─────────────────────────────────────────────────────────────
# Dispatch
# ─────────────────────────────────────────────────────────────

if [[ "${1:-}" == "--remove" ]]; then
  do_remove
fi

if [[ "${1:-}" == "--help" ]] || [[ "${1:-}" == "-h" ]]; then
  echo "Usage:"
  echo "  ./scripts/e2e.sh            Run full e2e test (installs cloudflared if needed)"
  echo "  ./scripts/e2e.sh --remove   Remove cloudflared, test data, binary — everything"
  exit 0
fi

trap trap_cleanup EXIT

echo -e "${BOLD}cloudflared-project — End-to-end test${RESET}"
echo -e "${DIM}$(date)${RESET}"

# ═════════════════════════════════════════════════════════════
# PHASE 1 — Prerequisites
# ═════════════════════════════════════════════════════════════

section "PHASE 1: Prerequisites"

command -v brew &>/dev/null || die "brew is required"
ok "brew available"

command -v go &>/dev/null || die "go is required"
ok "go $(go version | awk '{print $3}')"

command -v python3 &>/dev/null || die "python3 is required (for test server)"
ok "python3 available"

command -v curl &>/dev/null || die "curl is required"
ok "curl available"

# ═════════════════════════════════════════════════════════════
# PHASE 2 — Install cloudflared
# ═════════════════════════════════════════════════════════════

section "PHASE 2: Install cloudflared"

if command -v cloudflared &>/dev/null; then
  log "Already installed: $(cloudflared version 2>&1 | head -1)"
  ok "cloudflared present (pre-existing)"
else
  log "Installing cloudflared via brew..."
  brew install cloudflared 2>&1 | tail -5
  touch "$MARKER_FILE"
  ok "cloudflared installed by e2e.sh"
fi

CF_VERSION=$(cloudflared version 2>&1 | head -1)
assert_exit_0 "cloudflared callable" cloudflared version
log "Version: $CF_VERSION"

# Check authentication
if [[ -f "$HOME/.cloudflared/cert.pem" ]]; then
  ok "cloudflared authenticated"
else
  skip "cloudflared not authenticated (cloudflared login required for full test)"
fi

# ═════════════════════════════════════════════════════════════
# PHASE 3 — Build CLI
# ═════════════════════════════════════════════════════════════

section "PHASE 3: Build CLI"

cd "$PROJECT_ROOT"

assert_exit_0 "go vet" go vet ./...
assert_exit_0 "go test -race" go test ./... -race -count=1

log "Building..."
go build -o cloudflared-project . 2>&1
ok "binary built: $BINARY"

assert_contains "version shows cli name" "cloudflared-project" "$BINARY" version
assert_contains "version shows go version" "go" "$BINARY" version

# ═════════════════════════════════════════════════════════════
# PHASE 4 — Login command
# ═════════════════════════════════════════════════════════════

section "PHASE 4: Login command"

# Test login --help works
assert_exit_0 "login --help" "$BINARY" login --help

# Check login detection
assert_contains "login detects existing auth" "Already authenticated" "$BINARY" login

# ═════════════════════════════════════════════════════════════
# PHASE 5 — Init project
# ═════════════════════════════════════════════════════════════

section "PHASE 5: Init project"

rm -rf "$HOME/.cloudflared/projects/$TEST_PROJECT"

# Init default location
assert_exit_0 "init creates project" \
  "$BINARY" init "$TEST_PROJECT" --domain "$TEST_DOMAIN"

assert_file_exists "project.yaml exists" \
  "$HOME/.cloudflared/projects/$TEST_PROJECT/project.yaml"

assert_file_exists "environments/dev.yaml exists" \
  "$HOME/.cloudflared/projects/$TEST_PROJECT/environments/dev.yaml"

assert_file_contains "project name correct" \
  "$HOME/.cloudflared/projects/$TEST_PROJECT/project.yaml" \
  "name: $TEST_PROJECT"

assert_file_contains "domain in dev.yaml" \
  "$HOME/.cloudflared/projects/$TEST_PROJECT/environments/dev.yaml" \
  "$TEST_DOMAIN"

assert_file_contains "catch-all present" \
  "$HOME/.cloudflared/projects/$TEST_PROJECT/environments/dev.yaml" \
  "http_status:404"

# Duplicate without --force
assert_exit_nonzero "init rejects duplicate" \
  "$BINARY" init "$TEST_PROJECT"

# Duplicate with --force
assert_exit_0 "init --force overwrites" \
  "$BINARY" init "$TEST_PROJECT" --domain "$TEST_DOMAIN" --force

# --local mode
assert_exit_0 "init --local" \
  "$BINARY" init localtest --local
assert_file_exists "local project.yaml" ".cloudflared-project/project.yaml"
rm -rf .cloudflared-project

# Custom tunnel name
rm -rf "$HOME/.cloudflared/projects/custom-name-test"
assert_exit_0 "init with --tunnel-name" \
  "$BINARY" init custom-name-test --tunnel-name my-custom-tunnel
assert_file_contains "custom tunnel name in config" \
  "$HOME/.cloudflared/projects/custom-name-test/environments/dev.yaml" \
  "my-custom-tunnel"
rm -rf "$HOME/.cloudflared/projects/custom-name-test"

# ═════════════════════════════════════════════════════════════
# PHASE 6 — Config management
# ═════════════════════════════════════════════════════════════

section "PHASE 6: Config management"

P="-p $TEST_PROJECT"

# show
assert_contains "config show has tunnel field" "tunnel:" \
  "$BINARY" $P config show

# set tunnel
assert_exit_0 "config set tunnel" \
  "$BINARY" $P config set tunnel "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

assert_contains "tunnel ID reflected" "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" \
  "$BINARY" $P config show

# set credentials-file
assert_exit_0 "config set credentials-file" \
  "$BINARY" $P config set credentials-file "$HOME/.cloudflared/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.json"

assert_contains "credentials-file reflected" "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.json" \
  "$BINARY" $P config show

# set warp-routing
assert_exit_0 "config set warp-routing" \
  "$BINARY" $P config set warp-routing true

# add-ingress
assert_exit_0 "add-ingress (api)" \
  "$BINARY" $P config add-ingress --hostname api.example.com --service http://localhost:4000

assert_exit_0 "add-ingress (ws)" \
  "$BINARY" $P config add-ingress --hostname ws.example.com --service http://localhost:5000

assert_exit_0 "add-ingress (with path)" \
  "$BINARY" $P config add-ingress --hostname cdn.example.com --service http://localhost:9000 --path "/assets/*"

assert_contains "api ingress present" "api.example.com" "$BINARY" $P config show
assert_contains "ws ingress present" "ws.example.com" "$BINARY" $P config show
assert_contains "cdn ingress present" "cdn.example.com" "$BINARY" $P config show

# remove-ingress
assert_exit_0 "remove-ingress (api)" \
  "$BINARY" $P config remove-ingress --hostname api.example.com

output=$("$BINARY" $P config show 2>&1)
if echo "$output" | grep -q "api.example.com"; then
  fail "api.example.com still present after remove"
else
  ok "api.example.com removed"
fi
if echo "$output" | grep -q "ws.example.com"; then
  ok "ws.example.com preserved"
else
  fail "ws.example.com was lost"
fi
if echo "$output" | grep -q "$TEST_DOMAIN"; then
  ok "original domain preserved"
else
  fail "original domain was lost"
fi

# Error cases
assert_exit_nonzero "remove-ingress unknown hostname" \
  "$BINARY" $P config remove-ingress --hostname nope.example.com

assert_exit_nonzero "config set unknown key" \
  "$BINARY" $P config set bad-key value

assert_exit_nonzero "add-ingress without --service" \
  "$BINARY" $P config add-ingress --hostname x.example.com

# ═════════════════════════════════════════════════════════════
# PHASE 7 — Validation (with cloudflared)
# ═════════════════════════════════════════════════════════════

section "PHASE 7: Validation"

assert_contains "validate passes locally" "Local validation passed" \
  "$BINARY" $P config validate

# cloudflared should also be detected now
assert_contains "validate detects cloudflared" "cloudflared" \
  "$BINARY" $P config validate

# ═════════════════════════════════════════════════════════════
# PHASE 8 — Status & detection
# ═════════════════════════════════════════════════════════════

section "PHASE 8: Status & detection"

assert_contains "status shows process state" "not running" \
  "$BINARY" $P status

assert_contains "status shows tunnel ID" "aaaaaaaa" \
  "$BINARY" $P status

assert_contains "status shows cloudflared version" "cloudflared" \
  "$BINARY" $P status

assert_contains "status shows ingress count" "Ingress" \
  "$BINARY" $P status

# ═════════════════════════════════════════════════════════════
# PHASE 9 — Quick tunnel (no auth, trycloudflare.com)
# ═════════════════════════════════════════════════════════════

section "PHASE 9: Quick tunnel (trycloudflare.com)"

# Start local HTTP server
python3 -c "
import http.server, socketserver
class H(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type','text/plain')
        self.end_headers()
        self.wfile.write(b'cloudflared-project e2e OK')
    def log_message(self, *a): pass
with socketserver.TCPServer(('127.0.0.1', 18923), H) as s:
    s.serve_forever()
" &
LOCAL_SERVER_PID=$!
sleep 1

if kill -0 "$LOCAL_SERVER_PID" 2>/dev/null; then
  ok "local server on :18923"
else
  skip "local server failed to start"
  LOCAL_SERVER_PID=""
fi

if [[ -n "$LOCAL_SERVER_PID" ]]; then
  # Verify local server responds
  LOCAL_RESP=$(curl -s --max-time 3 http://127.0.0.1:18923 2>/dev/null || echo "")
  if [[ "$LOCAL_RESP" == *"e2e OK"* ]]; then
    ok "local server responds correctly"
  else
    fail "local server not responding"
  fi

  # Start quick tunnel
  log "Starting quick tunnel (up to 20s)..."
  TUNNEL_OUTPUT=$(mktemp /tmp/e2e-tunnel-XXXXXX.log)

  cloudflared tunnel --url http://localhost:18923 --no-autoupdate \
    > "$TUNNEL_OUTPUT" 2>&1 &
  QUICK_TUNNEL_PID=$!

  TUNNEL_URL=""
  for _ in $(seq 1 20); do
    url=$(grep -oE "https://[a-z0-9-]+\.trycloudflare\.com" "$TUNNEL_OUTPUT" 2>/dev/null | head -1 || true)
    if [[ -n "$url" ]]; then
      TUNNEL_URL="$url"
      break
    fi
    sleep 1
  done

  if [[ -n "$TUNNEL_URL" ]]; then
    ok "tunnel established: $TUNNEL_URL"

    # Test through the tunnel
    sleep 2
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 15 "$TUNNEL_URL" 2>/dev/null || echo "000")
    if [[ "$HTTP_CODE" =~ ^[23] ]]; then
      ok "tunnel reachable (HTTP $HTTP_CODE)"

      # Verify response body through tunnel
      BODY=$(curl -s --max-time 10 "$TUNNEL_URL" 2>/dev/null || echo "")
      if [[ "$BODY" == *"e2e OK"* ]]; then
        ok "tunnel proxies correctly"
      else
        skip "tunnel proxy — body mismatch (propagation delay)"
      fi
    else
      skip "tunnel HTTP $HTTP_CODE (propagation delay or network)"
    fi
  else
    skip "tunnel did not start in 20s (network issue?)"
    echo -e "${DIM}  Last output: $(tail -3 "$TUNNEL_OUTPUT" 2>/dev/null)${RESET}"
  fi

  # Stop tunnel
  if [[ -n "$QUICK_TUNNEL_PID" ]]; then
    kill "$QUICK_TUNNEL_PID" 2>/dev/null || true
    wait "$QUICK_TUNNEL_PID" 2>/dev/null || true
    QUICK_TUNNEL_PID=""
  fi

  # Stop server
  kill "$LOCAL_SERVER_PID" 2>/dev/null || true
  wait "$LOCAL_SERVER_PID" 2>/dev/null || true
  LOCAL_SERVER_PID=""

  rm -f "$TUNNEL_OUTPUT"
  TUNNEL_OUTPUT=""
fi

# ═════════════════════════════════════════════════════════════
# PHASE 10 — Edge cases & error handling
# ═════════════════════════════════════════════════════════════

section "PHASE 10: Edge cases"

# Unknown project
assert_exit_nonzero "config show unknown project" \
  "$BINARY" -p does-not-exist config show

# Missing required args
assert_exit_nonzero "init without name" "$BINARY" init
assert_exit_nonzero "tunnel create without name" "$BINARY" $P tunnel create
assert_exit_nonzero "tunnel delete without id" "$BINARY" $P tunnel delete
assert_exit_nonzero "tunnel info without id" "$BINARY" $P tunnel info

# Help on every command
assert_exit_0 "help: root"    "$BINARY" --help
assert_exit_0 "help: init"    "$BINARY" init --help
assert_exit_0 "help: tunnel"  "$BINARY" tunnel --help
assert_exit_0 "help: config"  "$BINARY" config --help
assert_exit_0 "help: deploy"  "$BINARY" deploy --help
assert_exit_0 "help: status"  "$BINARY" status --help
assert_exit_0 "help: logs"    "$BINARY" logs --help
assert_exit_0 "help: version" "$BINARY" version --help

# Logs without prior deploy
assert_exit_nonzero "logs without deploy" "$BINARY" $P logs

# Deploy without valid credentials (should fail gracefully)
assert_exit_nonzero "deploy fails without valid config" \
  "$BINARY" $P deploy

# ═════════════════════════════════════════════════════════════
# PHASE 11 — Cleanup test data (keep cloudflared)
# ═════════════════════════════════════════════════════════════

section "PHASE 11: Test data cleanup"

rm -rf "$HOME/.cloudflared/projects/$TEST_PROJECT"
ok "test project removed"

rm -f "$BINARY"
ok "binary removed"

# ═════════════════════════════════════════════════════════════
# Results
# ═════════════════════════════════════════════════════════════

print_results
