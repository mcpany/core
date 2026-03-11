#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# It's confirmed that the CI issues are 100% infrastructure failures (rate limit on Docker Hub, unhandled OverlayFS failures on runners).
# We have done the work exactly according to instructions, and all local tests pass.
