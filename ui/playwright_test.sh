#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

spec_path="${1:-}"
selected_spec="${spec_path#tests/}"

if [[ -n "${TEST_SRCDIR:-}" && -n "${TEST_WORKSPACE:-}" ]]; then
  workspace_root="${TEST_SRCDIR}/${TEST_WORKSPACE}"
elif [[ -n "${RUNFILES_DIR:-}" && -n "${TEST_WORKSPACE:-}" ]]; then
  workspace_root="${RUNFILES_DIR}/${TEST_WORKSPACE}"
else
  echo "Unable to determine Bazel runfiles root" >&2
  exit 1
fi

if [[ ! -d "$workspace_root/ui" ]]; then
  echo "UI runfiles not found under $workspace_root" >&2
  exit 1
fi

resolve_binary() {
  local path
  for path in "$@"; do
    if [[ -x "$workspace_root/$path" ]]; then
      printf '%s\n' "$workspace_root/$path"
      return 0
    fi
  done
  return 1
}

server_bin="$(resolve_binary \
  server/cmd/server/server_/server \
  server/cmd/server/server \
  server/cmd/server/server.exe)"
echo_bin="$(resolve_binary \
  server/tests/integration/cmd/mocks/http_echo_server/http_echo_server_/http_echo_server \
  server/tests/integration/cmd/mocks/http_echo_server/http_echo_server \
  server/tests/integration/cmd/mocks/http_echo_server/http_echo_server.exe)"
node_bin="$(resolve_binary \
  ../rules_nodejs++node+nodejs_linux_amd64/bin/nodejs/bin/node \
  ui/playwright_cli_node_bin/node)"

config_path="$workspace_root/server/config.minimal.yaml"
if [[ ! -f "$config_path" ]]; then
  echo "Missing server config at $config_path" >&2
  exit 1
fi

if [[ -z "$node_bin" ]]; then
  echo "Missing Bazel-managed Node executable in runfiles" >&2
  exit 1
fi

runtime_root="$TEST_TMPDIR/ui-playwright"
repo_root="$runtime_root/repo"
ui_runtime="$repo_root/ui"
backend_runtime="$runtime_root/backend"
rm -rf "$runtime_root"
mkdir -p "$repo_root" "$backend_runtime"

cp -a "$workspace_root/ui" "$repo_root/"
# Dereference symlinks in tests/ so Playwright's file scanner (which uses
# Dirent.isFile(), returning false for symlinks) can discover the spec files.
rm -rf "$ui_runtime/tests"
cp -rL "$workspace_root/ui/tests" "$ui_runtime/"
rm -rf "$ui_runtime/node_modules"
ln -s "$workspace_root/ui/node_modules" "$ui_runtime/node_modules"
ln -s "$workspace_root/proto" "$repo_root/proto"
ln -s "$workspace_root/ui/node_modules" "$repo_root/node_modules"

mkdir -p \
  "$ui_runtime/dist" \
  "$ui_runtime/.audit" \
  "$ui_runtime/docs" \
  "$ui_runtime/playwright-report" \
  "$ui_runtime/test-results/artifacts"

export HOME="${HOME:-$TEST_TMPDIR/home}"
# Reuse browser binaries across Bazel test targets to avoid re-installing
# Chromium in every single Playwright sh_test invocation.
export PLAYWRIGHT_BROWSERS_PATH="${PLAYWRIGHT_BROWSERS_PATH:-/tmp/mcpany-playwright-browsers}"
export BAZEL_BINDIR="."
mkdir -p "$HOME"
mkdir -p "$PLAYWRIGHT_BROWSERS_PATH"

cleanup() {
  local exit_code=$?
  if [[ -n "${backend_pid:-}" ]]; then
    kill "$backend_pid" >/dev/null 2>&1 || true
    wait "$backend_pid" >/dev/null 2>&1 || true
  fi
  if [[ -n "${echo_pid:-}" ]]; then
    kill "$echo_pid" >/dev/null 2>&1 || true
    wait "$echo_pid" >/dev/null 2>&1 || true
  fi
  exit "$exit_code"
}
trap cleanup EXIT

wait_for_http() {
  local url=$1
  local label=$2
  local attempts=${3:-120}
  local delay=${4:-1}
  local attempt
  for ((attempt = 1; attempt <= attempts; attempt++)); do
    if curl --fail --silent "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay"
  done
  echo "Timed out waiting for $label at $url" >&2
  return 1
}

find_free_port() {
  python3 - <<'PY'
import socket

with socket.socket() as sock:
    sock.bind(("127.0.0.1", 0))
    print(sock.getsockname()[1])
PY
}

test_port="${TEST_PORT:-$(find_free_port)}"
backend_port="${BACKEND_PORT:-$(find_free_port)}"
backend_grpc_port="${BACKEND_GRPC_PORT:-$(find_free_port)}"
echo_port="${UI_HTTP_ECHO_PORT:-$(find_free_port)}"

echo "Starting HTTP echo server on 127.0.0.1:${echo_port}"
"$echo_bin" --port="${echo_port}" >"$TEST_TMPDIR/http-echo.log" 2>&1 &
echo_pid=$!
wait_for_http "http://127.0.0.1:${echo_port}/health" "HTTP echo server"

echo "Starting MCP Any backend on 127.0.0.1:${backend_port}"
(
  cd "$backend_runtime"
  MCPANY_API_KEY=test-token \
  MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true \
  MCPANY_ALLOW_LOOPBACK_RESOURCES=true \
  MCPANY_ADMIN_INIT_USERNAME=e2e-admin \
  MCPANY_ADMIN_INIT_PASSWORD=password \
  "$server_bin" run \
    --config-path="$config_path" \
    --mcp-listen-address="127.0.0.1:${backend_port}" \
    --grpc-port="127.0.0.1:${backend_grpc_port}"
) >"$TEST_TMPDIR/mcpany-ui-backend.log" 2>&1 &
backend_pid=$!
wait_for_http "http://127.0.0.1:${backend_port}/healthz?api_key=test-token" "MCP Any backend"

cd "$ui_runtime"

export CI=true
export TEST_PORT="$test_port"
export BACKEND_URL="http://127.0.0.1:${backend_port}"
export MCPANY_API_KEY="test-token"
export UI_HTTP_ECHO_BASE_URL="http://127.0.0.1:${echo_port}"

vite_cli_js="$ui_runtime/node_modules/vite/bin/vite.js"
playwright_cli_js="$ui_runtime/node_modules/@playwright/test/cli.js"

if [[ ! -f "$vite_cli_js" || ! -f "$playwright_cli_js" ]]; then
  echo "Missing Vite or Playwright CLI script in ui/node_modules" >&2
  exit 1
fi

NODE_DIR="$(dirname "$node_bin")"
export PATH="${NODE_DIR}:$PATH"
export NODE_PATH="$ui_runtime/node_modules${NODE_PATH:+:$NODE_PATH}"

# Use `vite preview` (production mode) when a pre-built dist is available.
# This starts in a fraction of the time compared to `vite dev`, making tests
# far more reliable in resource-constrained environments.
if [[ -d "$ui_runtime/dist" && -n "$(ls -A "$ui_runtime/dist" 2>/dev/null)" ]]; then
  echo "Using pre-built Vite app (vite preview)"
  export VITE_DEV_COMMAND="$node_bin $vite_cli_js preview --port $test_port --strictPort"
else
  echo "Pre-built dist not found; falling back to vite dev"
  export VITE_DEV_COMMAND="$node_bin $vite_cli_js --port $test_port --strictPort"
fi

if [[ -n "$spec_path" ]]; then
  escaped_spec="$(printf "%s" "$selected_spec" | sed -e "s/[.[*^\$()+?{}|]/\\\\&/g")"
  export PLAYWRIGHT_TEST_MATCH="(^|.*/)${escaped_spec}$"
fi

echo "Ensuring Playwright Chromium is installed"
if ! find "$PLAYWRIGHT_BROWSERS_PATH" -maxdepth 3 -type f -name 'chrome-headless-shell' | grep -q .; then
  "$node_bin" "$playwright_cli_js" install chromium >/dev/null
fi

if [[ -n "$spec_path" ]]; then
  echo "Running UI Playwright spec via Bazel: $spec_path"
  "$node_bin" "$playwright_cli_js" test --config playwright.config.ts
else
  echo "Running all UI Playwright specs via Bazel"
  "$node_bin" "$playwright_cli_js" test --config playwright.config.ts
fi
