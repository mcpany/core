#!/bin/bash

# Simple script to figure out exactly what is failing in the 'make lint' step
echo "--- Running buildifier ---"
/app/build/env/bin/bazelisk run //:lint 2>&1 | tee buildifier_out.log
echo "Buildifier returned: ${PIPESTATUS[0]}"

echo "--- Running golangci-lint directly ---"
cd server
./../build/env/bin/golangci-lint run --timeout 20m --fix ./cmd/... ./pkg/... ./tests/... ./examples/... 2>&1 | tee ../golangci_out.log
echo "golangci-lint returned: ${PIPESTATUS[0]}"
cd ..

echo "--- Running npm lint directly ---"
cd ui
npm run lint 2>&1 | tee ../npm_lint_out.log
echo "npm lint returned: ${PIPESTATUS[0]}"
cd ..
