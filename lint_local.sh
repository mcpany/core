#!/bin/bash
cd server
export PATH="/app/build/env/bin:$PATH"
golangci-lint run ./pkg/middleware/...
