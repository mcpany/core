#!/bin/bash
cd server
gofumpt -l -w cmd/server/main_test.go cmd/server/main.go
