#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


read -r input
message=$(echo "$input" | jq -r .message)
echo "{\"message\": \"$message\"}"
