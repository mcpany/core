#!/bin/bash
export GO111MODULE=on
export GOWORK=off
cd server
cat << 'CONFIG_EOF' > .golangci.yml
run:
  timeout: 5m
linters:
  enable:
    - revive
    - govet
issues:
  exclude-dirs:
    - vendor
CONFIG_EOF
