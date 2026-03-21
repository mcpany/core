#!/bin/bash
export PATH=/home/jules/go/bin:$PATH
cd server
golangci-lint run ./...
