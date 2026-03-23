#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Determine project root
if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
fi
cd "$PROJECT_ROOT"

echo "==> Running lint automation..."

# Ensure we are using Go 1.26.1 as required by the project
export GOTOOLCHAIN=go1.26.1
echo "    Go version: $(go version)"

# Build/locate a compatible golangci-lint
LINT_VERSION="v1.64.5"
# Use a dedicated binary path to avoid version conflicts with system-installed linters
LINT_BIN="$PROJECT_ROOT/build/env/bin/golangci-lint"
mkdir -p "$(dirname "$LINT_BIN")"

if [[ ! -x "$LINT_BIN" ]] || ! ("$LINT_BIN" --version | grep -q "go1.26"); then
    echo "    Building golangci-lint ${LINT_VERSION} from source with Go 1.26.1..."
    # Install to a temporary location then move to avoid partial binary execution
    TMP_GOBIN=$(mktemp -d)
    GOTOOLCHAIN=go1.26.1 GOBIN="$TMP_GOBIN" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
    mv "$TMP_GOBIN/golangci-lint" "$LINT_BIN"
    rm -rf "$TMP_GOBIN"
fi

echo "    Using linter: $($LINT_BIN --version)"

# Lint each module independently to avoid workspace/module boundary issues (GOWORK=off)
export GOWORK=off
MODULES=("server" "proto" "k8s/operator")

# Absolute path to config
CONFIG_PATH="$PROJECT_ROOT/server/.golangci.yml"

for mod in "${MODULES[@]}"; do
    if [ -d "$mod" ]; then
        echo "    Linting module: $mod"
        (cd "$mod" && "$LINT_BIN" run --timeout 20m --fix --config "$CONFIG_PATH" ./...)
    fi
done

# Run Buildifier and Gazelle if available
if command -v buildifier >/dev/null 2>&1; then
    echo "    Running Buildifier..."
    find . -name "BUILD" -o -name "BUILD.bazel" -o -name "*.bzl" -not -path "./build/*" -exec buildifier {} +
fi

if command -v gazelle >/dev/null 2>&1; then
    echo "    Running Gazelle..."
    gazelle -repo_root="$PROJECT_ROOT"
fi

echo "==> Lint complete."
