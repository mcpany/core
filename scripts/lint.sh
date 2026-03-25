#!/usr/bin/env bash
set -e
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"

echo "==> Running Buildifier..."
BUILDIFIER_BIN="/tmp/buildifier"
if [[ -x "$BUILDIFIER_BIN" ]]; then
    $BUILDIFIER_BIN -mode=check $(find . -path "./build" -prune -o -path "./bazel-*" -prune -o -path "*/node_modules" -prune -o \( -name "BUILD" -o -name "BUILD.bazel" -o -name "WORKSPACE" -o -name "*.bzl" \) -type f -print)
    echo "    Buildifier OK."
else
    echo "    Warning: buildifier not found."
fi

echo "==> Running golangci-lint..."
GOLANGCI_LINT_BIN="/tmp/golangci-lint"
if [[ ! -x "$GOLANGCI_LINT_BIN" ]]; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /tmp v1.64.5
fi

# Run lint for each module in go.work
echo "Checking server..."
cd server && $GOLANGCI_LINT_BIN run --timeout 10m --config .golangci.yml ./...
cd ..

echo "Checking k8s/operator..."
cd k8s/operator && $GOLANGCI_LINT_BIN run --timeout 10m ./...
cd ../..

echo "Checking root..."
$GOLANGCI_LINT_BIN run --timeout 10m --config server/.golangci.yml ./...

echo "==> Lint complete."
