#!/bin/bash
export PATH=$PATH:~/go/bin
cd server
golangci-lint run --out-format=line-number --disable-all --enable=gofmt --go 1.24 ./... > lint.log 2>&1
cat lint.log
