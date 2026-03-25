#!/usr/bin/env bash
set -e
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"
export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"
LINT_DIR="${PROJECT_ROOT}/build/env/bin"
mkdir -p "$LINT_DIR"
LINT_BIN="$LINT_DIR/golangci-lint"
if [[ ! -x "$LINT_BIN" ]] || ! "$LINT_BIN" --version | grep -q "${LINT_VERSION}" || ! "$LINT_BIN" --version | grep -q "go1.26"; then
    GOBIN="$LINT_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi
export GOWORK="${PROJECT_ROOT}/go.work"
go work sync
"$LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml ./...
