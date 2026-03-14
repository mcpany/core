#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# This script is intended to be run via Bazel: bazel run //:lint
# It wraps the existing linting logic from the Makefile.

# --- begin runfiles.bash initialization v3 ---
# Loads the Bazel bash runfiles library (only available when running via bazel run).
# Silently skipped when running outside of Bazel (e.g. via make lint directly).
# shellcheck disable=SC1090
set +e
_f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck disable=SC2086
source "${RUNFILES_DIR:-/dev/null}/${_f}" 2>/dev/null ||
    source "$(grep -sm1 "^${_f} " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null ||
    true
unset _f
set -e
# --- end runfiles.bash initialization v3 ---

if [ -n "$BUILD_WORKSPACE_DIRECTORY" ]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT=$(bazel info workspace)
fi
cd "$PROJECT_ROOT"

echo "Running Gazelle..."
bazel run //:gazelle

echo "Running Buildifier..."
bazel run //:buildifier -- -r .

echo "Running golangci-lint..."
# Prefer the golangci-lint binary bundled by Bazel (via runfiles), so that
# the CI runner needs no pre-installed tooling beyond Bazelisk.
GOLANGCI_LINT_BIN=""
if declare -f rlocation >/dev/null 2>&1; then
    for _plat in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64; do
        _bin=$(rlocation "golangci_lint_${_plat}/golangci-lint" 2>/dev/null || true)
        if [ -x "${_bin:-}" ]; then
            GOLANGCI_LINT_BIN="$_bin"
            break
        fi
    done
    unset _plat _bin
fi

# Fall back to the path used by "make prepare" or a globally installed binary.
if [ -z "$GOLANGCI_LINT_BIN" ]; then
    GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN_FALLBACK:-build/env/bin/golangci-lint}"
fi
if ! [ -x "$GOLANGCI_LINT_BIN" ]; then
    GOLANGCI_LINT_BIN=$(which golangci-lint 2>/dev/null || true)
fi

if [ -z "$GOLANGCI_LINT_BIN" ]; then
    echo "Error: golangci-lint not found. Run 'make prepare' or use 'bazel run //:lint'."
    exit 1
fi

"$GOLANGCI_LINT_BIN" run --timeout 20m --fix \
    ./server/cmd/... \
    ./server/pkg/... \
    ./server/tests/... \
    ./server/examples/...

echo "Running pre-commit..."
if command -v pre-commit >/dev/null 2>&1; then
    pre-commit run --config server/.pre-commit-config.yaml --all-files
else
    echo "Warning: pre-commit not found. Skipping."
fi

