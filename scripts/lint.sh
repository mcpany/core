#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Load runfiles library if possible
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
        case "$name" in
            buildifier) bin="$(rlocation buildifier_prebuilt/buildifier)" ;;
            golangci-lint) bin="$(rlocation golangci_lint_bin/golangci-lint)" ;;
        esac
    fi
    if [[ -z "$bin" ]]; then
        bin="$(command -v "${name}" 2>/dev/null || true)"
        if [[ -z "$bin" && -x "${PROJECT_ROOT}/build/env/bin/${name}" ]]; then
            bin="${PROJECT_ROOT}/build/env/bin/${name}"
        fi
        if [[ -z "$bin" && -x "/tmp/${name}" ]]; then
            bin="/tmp/${name}"
        fi
    fi
    echo "$bin"
}

echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -n "$BUILDIFIER_BIN" && -x "$BUILDIFIER_BIN" ]]; then
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
GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"

if [[ ! -x "$GOLANGCI_LINT_BIN" ]]; then
    echo "    Installing golangci-lint v1.64.5..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /tmp v1.64.5
    GOLANGCI_LINT_BIN="/tmp/golangci-lint"
fi

# Run lint for each module in go.work
echo "    Checking modules..."
LINT_FAILED=0
# Parse go.work to get directories
MODULES=$(grep -E '^\s+\./' go.work | sed 's/^\s+//')
if [ -z "$MODULES" ]; then
    # Fallback to manual list if parsing fails
    MODULES="./ ./k8s/operator ./server ./server/docs/features/webhooks/examples/block_rm ./server/docs/features/webhooks/examples/html_to_md ./server/examples/upstream_service_demo/grpc/client ./server/examples/upstream_service_demo/grpc/greeter_server"
fi

for dir in $MODULES; do
    echo "    Checking $dir..."
    (
        cd "$dir"
        CONFIG=""
        if [ -f ".golangci.yml" ]; then
            CONFIG=".golangci.yml"
        elif [ -f "$PROJECT_ROOT/server/.golangci.yml" ]; then
            CONFIG="$PROJECT_ROOT/server/.golangci.yml"
        fi

        if [ -n "$CONFIG" ]; then
            "$GOLANGCI_LINT_BIN" run --timeout 10m --config "$CONFIG" ./...
        else
            "$GOLANGCI_LINT_BIN" run --timeout 10m ./...
        fi
    ) || LINT_FAILED=1
done

if [ "$LINT_FAILED" -ne 0 ]; then
    echo "==> Lint FAILED."
    false
fi

echo "==> Lint complete."
