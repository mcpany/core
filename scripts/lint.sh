#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# This script is intended to be run via Bazel: bazel run //:lint
# All linting tools are provided as Bazel data dependencies so that
# no non-Bazel installs are required.  Results are cached by Bazel's
# remote/disk cache when run with --config=remote or --disk_cache.

# ---------------------------------------------------------------------------
# Bazel runfiles helper – lets us locate data-dep binaries via rlocation().
# ---------------------------------------------------------------------------
# shellcheck source=/dev/null
if [[ -f "${RUNFILES_DIR:-/dev/null}/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    source "${RUNFILES_DIR}/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -f "${0}.runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    source "${0}.runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -f "$(dirname "${BASH_SOURCE[0]}")/runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    source "$(dirname "${BASH_SOURCE[0]}")/runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash"
fi

# ---------------------------------------------------------------------------
# Determine project root.  When run via `bazel run`, BUILD_WORKSPACE_DIRECTORY
# is set to the workspace root (the real source tree, not the sandbox).
# ---------------------------------------------------------------------------
if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
fi
cd "$PROJECT_ROOT"

# ---------------------------------------------------------------------------
# Helper: locate a binary from Bazel runfiles, then fall back to $PATH.
# Usage: find_tool <binary-name>
# ---------------------------------------------------------------------------
find_tool() {
    local name="$1"
    local bin=""

    # 1. Search the runfiles tree (available when run under `bazel run`).
    if [[ -d "${RUNFILES_DIR:-}" ]]; then
        bin="$(find "${RUNFILES_DIR}" -name "${name}" -type f 2>/dev/null | head -1 || true)"
    fi

    # 2. Fall back to $PATH.
    if [[ -z "$bin" || ! -x "$bin" ]]; then
        bin="$(command -v "${name}" 2>/dev/null || true)"
    fi

    echo "$bin"
}

# ---------------------------------------------------------------------------
# 1. Buildifier – formats/lints Bazel BUILD files.
#    Binary supplied via data dep @buildifier_prebuilt//:buildifier.
# ---------------------------------------------------------------------------
echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -z "$BUILDIFIER_BIN" || ! -x "$BUILDIFIER_BIN" ]]; then
    echo "ERROR: buildifier not found. It should be provided as a Bazel data dep." >&2
    exit 1
fi
"$BUILDIFIER_BIN" -r .
echo "    Buildifier OK."

# ---------------------------------------------------------------------------
# 2. Gazelle – keeps Go BUILD targets in sync with Go source files.
#    Binary supplied via data dep @gazelle//:gazelle.
# ---------------------------------------------------------------------------
echo "==> Running Gazelle..."
GAZELLE_BIN="$(find_tool gazelle)"
if [[ -n "$GAZELLE_BIN" && -x "$GAZELLE_BIN" ]]; then
    "$GAZELLE_BIN" -repo_root="$PROJECT_ROOT"
    echo "    Gazelle OK."
else
    echo "    Warning: gazelle binary not found in runfiles – skipping."
    echo "    To update BUILD files manually, run: bazel run //:gazelle"
fi

# ---------------------------------------------------------------------------
# 3. golangci-lint – comprehensive Go static analysis.
#    Prefer the Bazel-managed binary (data dep :golangci_lint_bin), then
#    the path set by $GOLANGCI_LINT_BIN, then a local install, then
#    build/env/bin/ (populated by `make prepare`).
# ---------------------------------------------------------------------------
echo "==> Running golangci-lint..."
if [[ -z "${GOLANGCI_LINT_BIN:-}" ]]; then
    GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"
fi
if [[ -z "${GOLANGCI_LINT_BIN:-}" || ! -x "${GOLANGCI_LINT_BIN}" ]]; then
    GOLANGCI_LINT_BIN="build/env/bin/golangci-lint"
fi

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    "$GOLANGCI_LINT_BIN" run --timeout 20m --fix \
        ./server/cmd/... ./server/pkg/... ./server/tests/... ./server/examples/...
    echo "    golangci-lint OK."
else
    echo "    Warning: golangci-lint not found."
    echo "    To install via Bazel, ensure the :golangci_lint_bin data dep is present."
    echo "    Or run 'make prepare' to install it locally."
fi

echo "==> Lint complete."
