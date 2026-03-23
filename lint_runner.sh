#!/bin/bash
cd server
GOGC=20 ../build/env/bin/golangci-lint run --concurrency 1 ./... > lint_output.txt 2>&1
