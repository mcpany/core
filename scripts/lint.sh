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
    if [[ $(type -t rlocation) == function ]]; then
        # When running under Bazel, use rlocation to find tools.
        case "$name" in
            buildifier) bin="$(rlocation buildifier_prebuilt/buildifier)" ;;
            golangci-lint) bin="$(rlocation golangci_lint_bin/golangci-lint)" ;;
        esac
    fi
    if [[ -z "$bin" ]]; then
        # Fallback to local build env or PATH
        bin="$(command -v "${name}" 2>/dev/null || true)"
        if [[ -z "$bin" && -x "${PROJECT_ROOT}/build/env/bin/${name}" ]]; then
            bin="${PROJECT_ROOT}/build/env/bin/${name}"
        fi
    fi
    echo "$bin"
}

echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -n "$BUILDIFIER_BIN" && -x "$BUILDIFIER_BIN" ]]; then
    # Use -prune to efficiently skip directories that should not be searched.
    BUILDIFIER_FILES=$(find . \
        -path "./build" -prune -o \
        -path "./bazel-*" -prune -o \
        -path "*/node_modules" -prune -o \
        \( -name "BUILD" -o -name "BUILD.bazel" -o -name "WORKSPACE" -o -name "*.bzl" \) -type f -print)

    if [[ -n "$BUILDIFIER_FILES" ]]; then
        "$BUILDIFIER_BIN" -mode=check $BUILDIFIER_FILES
    fi
    echo "    Buildifier OK."
else
    echo "    Warning: buildifier not found."
fi

echo "==> Running golangci-lint..."
if [[ -z "${GOLANGCI_LINT_BIN:-}" ]]; then
    GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"
fi

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    # Run golangci-lint on the workspace.
    # Note: If Go version > 1.24, golangci-lint v1.64.5 might fail due to toolchain mismatch.
    "$GOLANGCI_LINT_BIN" run --timeout 20m --config server/.golangci.yml ./...
    echo "    golangci-lint OK."
else
    echo "    Warning: golangci-lint not found (skipping Go linting)."
fi

echo "==> Lint complete."
