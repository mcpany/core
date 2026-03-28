#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

if [[ -f "${RUNFILES_DIR:-/dev/null}/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    source "${RUNFILES_DIR}/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -f "${0}.runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    source "${0}.runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -f "$(dirname "${BASH_SOURCE[0]}")/runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    source "$(dirname "${BASH_SOURCE[0]}")/runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash"
fi

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
fi
cd "$PROJECT_ROOT"

find_tool() {
    local name="$1"
    local bin=""
    local search_dirs=()
    [[ -d "${RUNFILES_DIR:-}" ]] && search_dirs+=("${RUNFILES_DIR}")
    [[ -d "${0}.runfiles" ]] && search_dirs+=("${0}.runfiles")
    [[ -d "${BASH_SOURCE[0]}.runfiles" ]] && search_dirs+=("${BASH_SOURCE[0]}.runfiles")
    for dir in "${search_dirs[@]}"; do
        bin="$(find -L "${dir}" -name "${name}" \( -type f -o -type l \) 2>/dev/null | head -1 || true)"
        if [[ -n "$bin" && -x "$bin" ]]; then
            echo "$bin"
            return 0
        fi
    done
    bin="$(command -v "${name}" 2>/dev/null || true)"
    echo "$bin"
}

echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -x "$BUILDIFIER_BIN" ]]; then
    buildifier_files=($(find . -not -path '*/.*' \( -name 'BUILD' -o -name 'WORKSPACE' -o -name '*.bzl' \) -type f))
    if [[ ${#buildifier_files[@]} -gt 0 ]]; then
        "$BUILDIFIER_BIN" "${buildifier_files[@]}"
    fi
    echo "    Buildifier OK."
else
    echo "    Warning: buildifier not found - skipping."
fi

echo "==> Running golangci-lint..."
# Prioritize PATH or GOLANGCI_LINT_BIN, then GOPATH, then find_tool
GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN:-$(command -v golangci-lint || true)}"
if [[ ! -x "$GOLANGCI_LINT_BIN" ]]; then
    GOLANGCI_LINT_BIN="$(go env GOPATH)/bin/golangci-lint"
fi
if [[ ! -x "$GOLANGCI_LINT_BIN" ]]; then
    GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"
fi

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    "$GOLANGCI_LINT_BIN" run --timeout 20m --fix ./server/cmd/... ./server/pkg/... ./server/tests/... ./server/examples/...
    echo "    golangci-lint OK."
else
    echo "    Error: golangci-lint not found."
    return 1
fi

echo "==> Running Docstring Audit..."
python3 server/tools/audit_docs.py .
echo "    Docstring Audit OK."

echo "==> Lint complete."
