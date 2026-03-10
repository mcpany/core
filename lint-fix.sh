#!/bin/bash
set -e

# Run the fix scripts directly to address lint errors

# Fix bazel BUILD file issues
bazel mod tidy || true
bazel run //:buildifier || true
bazel run //:gazelle || true

# Run golangci-lint with auto-fix
cd server
golangci-lint run --fix || true
cd ..
