#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# This script is intended to be run via Bazel: bazel run //:lint
# All linting tools are provided as Bazel data dependencies so that
# no non-Bazel installs are required.  Results are cached by Bazel's
# remote/disk cache when run with --config=remote or --disk_cache.

# ---------------------------------------------------------------------------
# Determine project root.  When run via `bazel run`, BUILD_WORKSPACE_DIRECTORY
# is set to the workspace root (the real source tree, not the sandbox).
# ---------------------------------------------------------------------------
if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
fi
cd "$PROJECT_ROOT" || { echo "Could not cd to project root"; exit 1; }

# ---------------------------------------------------------------------------
# Helper: locate a binary from Bazel runfiles, then fall back to $PATH.
# Usage: find_tool <binary-name>
# ---------------------------------------------------------------------------
find_tool() {
    local name="$1"
    local bin=""
    local search_dirs=()

    # Collect runfiles search roots:
    # 1. $RUNFILES_DIR (set when use_bash_launcher is used)
    [[ -d "${RUNFILES_DIR:-}" ]] && search_dirs+=("${RUNFILES_DIR}")
    # 2. $0.runfiles (set by `bazel run` when the binary is run directly)
    [[ -d "${0}.runfiles" ]] && search_dirs+=("${0}.runfiles")

    for dir in "${search_dirs[@]}"; do
        bin="$(find -L "${dir}" -name "${name}" \( -type f -o -type l \) 2>/dev/null | head -1 || true)"
        if [[ -n "$bin" && -x "$bin" ]]; then
            echo "$bin"
            return 0
        fi
    done

    # 3. Fall back to $PATH.
    bin="$(command -v "${name}" 2>/dev/null || true)"
    echo "$bin"
}

# ---------------------------------------------------------------------------
# 1. Buildifier – formats/lints Bazel BUILD files.
# ---------------------------------------------------------------------------
echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -z "$BUILDIFIER_BIN" || ! -x "$BUILDIFIER_BIN" ]]; then
    echo "ERROR: buildifier not found. It should be provided as a Bazel data dep." >&2
    exit 1
fi

# Collect Bazel BUILD / .bzl / WORKSPACE files, excluding caches and symlinks.
# Use while read loop to handle filenames correctly
while IFS= read -r f; do
    "$BUILDIFIER_BIN" "$f"
done < <(find . \
    -not \( \
        -path './build/*' \
        -o -path './bazel-*' \
        -o -path './node_modules/*' \
        -o -path './.git/*' \
        -o -path './ui/node_modules/*' \
        -o -path './server/node_modules/*' \
    \) \
    \( \
        -name 'BUILD' \
        -o -name 'BUILD.bazel' \
        -o -name 'WORKSPACE' \
        -o -name 'WORKSPACE.bazel' \
        -o -name '*.bzl' \
    \) \
    -type f \
    2>/dev/null)

echo "    Buildifier OK."

# ---------------------------------------------------------------------------
# 2. Gazelle – keeps Go BUILD targets in sync with Go source files.
# ---------------------------------------------------------------------------
echo "==> Running Gazelle..."
GAZELLE_BIN="$(find_tool gazelle)"
if [[ -n "$GAZELLE_BIN" && -x "$GAZELLE_BIN" ]]; then
    "$GAZELLE_BIN" -repo_root="$PROJECT_ROOT"
    echo "    Gazelle OK."
else
    echo "    Warning: gazelle binary not found in runfiles – skipping."
fi

echo "==> Lint complete."
