#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
fi
cd "$PROJECT_ROOT"

echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(command -v buildifier 2>/dev/null || true)"
if [[ -z "$BUILDIFIER_BIN" ]]; then
    # Try finding it in runfiles if it's not in PATH
    BUILDIFIER_BIN="$(find -L . -name buildifier -type f -executable 2>/dev/null | head -1)"
fi
if [[ -n "$BUILDIFIER_BIN" && -x "$BUILDIFIER_BIN" ]]; then
    buildifier_files=(
        $(find . \
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
    )
    if [[ ${#buildifier_files[@]} -gt 0 ]]; then
        "$BUILDIFIER_BIN" "${buildifier_files[@]}"
    fi
    echo "    Buildifier OK."
else
    echo "    Warning: buildifier not found."
fi


echo "==> Running Gazelle..."
GAZELLE_BIN="$(command -v gazelle 2>/dev/null || true)"
if [[ -n "$GAZELLE_BIN" && -x "$GAZELLE_BIN" ]]; then
    "$GAZELLE_BIN" -repo_root="$PROJECT_ROOT"
    echo "    Gazelle OK."
else
    echo "    Warning: gazelle binary not found."
fi

echo "==> Running golangci-lint..."
GOLANGCI_LINT_BIN="$(command -v golangci-lint 2>/dev/null || true)"
if [[ -n "$GOLANGCI_LINT_BIN" && -x "$GOLANGCI_LINT_BIN" ]]; then
    cd server
    "$GOLANGCI_LINT_BIN" run --timeout 20m --fix ./cmd/... ./pkg/... ./tests/... ./examples/...
    echo "    golangci-lint OK."
    cd ..
else
    echo "    Warning: golangci-lint not found (skipping Go linting)."
fi

echo "==> Lint complete."
