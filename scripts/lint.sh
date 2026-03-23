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

echo "==> Running golangci-lint via go run (guaranteed alignment)..."
export GOTOOLCHAIN=go1.26.1
# Disable GOWORK for the linter run to isolate module analysis
export GOWORK=off

LINT_VERSION="v1.64.5"
LINT_CMD="go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}"

# Define absolute path to config
CONFIG_PATH="$(pwd)/server/.golangci.yml"

# Lint specific modules
MODULES=("server" "proto" "k8s/operator")

for mod in "${MODULES[@]}"; do
    if [ -d "$mod" ]; then
        echo "    Linting module: $mod"
        (cd "$mod" && $LINT_CMD run --timeout 20m --fix --config "$CONFIG_PATH" ./...)
    fi
done

echo "    golangci-lint OK."
echo "==> Lint complete."
