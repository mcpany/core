#!/bin/bash
export BAZEL_BINDIR="."
export PATH=$PATH:/usr/local/go/bin
cd proto
# Simulate what ts_proto_rules.bzl does or just let npm run lint ignore it if it can
cd ../ui
npm install
npm run lint
