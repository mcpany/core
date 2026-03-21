#!/bin/bash
cd server
go mod tidy
bazel run //:gazelle
bazel run //:gazelle-update-repos
