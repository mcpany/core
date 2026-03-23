#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/callDef := configv1.HttpCallDefinition_builder{/_ = configv1.HttpCallDefinition_builder{/' server/tests/integration/e2e_helpers.go
