#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e
bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=repositories.bzl%go_repositories
