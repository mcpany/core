#!/bin/bash
cd server

# Because errcheck won't run cleanly due to some go build errors when not built inside bazel
# we use our custom AST parser. Let's see if we missed anything in handleTools
go run ../lint_fix.go
