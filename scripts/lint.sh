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

echo "==> Running lint automation..."

export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"
LINT_DIR="$PROJECT_ROOT/build/linter"
mkdir -p "$LINT_DIR"

if [[ ! -x "$LINT_DIR/golangci-lint" ]]; then
    echo "    Building golangci-lint ${LINT_VERSION} from source..."
    GOWORK=off GOTOOLCHAIN=go1.26.1 GOBIN="$LINT_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi

LINT_BIN="$LINT_DIR/golangci-lint"
echo "    Using linter: $($LINT_BIN --version)"

# To avoid "directory is outside module roots" error, we must NOT use ./... from root if go.work is active but we want module-level context,
# OR we must ensure go.work is handled correctly.
# golangci-lint supports workspaces, but sometimes it's finicky in CI.
# Let's try running it on each module with GOWORK=off to be safe.

export GOWORK=off
MODULES=("server" "proto" "k8s/operator")

for mod in "${MODULES[@]}"; do
    if [ -d "$mod" ]; then
        echo "    Linting module: $mod"
        # We must use absolute path to config or relative from the module dir
        (cd "$mod" && "$LINT_BIN" run --timeout 20m --fix --config ../server/.golangci.yml ./...)
    fi
done

echo "==> Lint complete."
