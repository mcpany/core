#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/configv1\.VectorUpstreamService_builder{/configv1.VectorUpstreamService_builder{/g' server/pkg/util/secrets_sanitizer_test.go
