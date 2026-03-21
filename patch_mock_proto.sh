#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/export interface CallPolicyRule/\/** Mock CallPolicyRule *\/\nexport interface CallPolicyRule/' ui/src/mocks/proto/mock-proto.ts
sed -i 's/export interface ExportPolicy/\/** Mock ExportPolicy *\/\nexport interface ExportPolicy/' ui/src/mocks/proto/mock-proto.ts
sed -i 's/export interface ExportRule/\/** Mock ExportRule *\/\nexport interface ExportRule/' ui/src/mocks/proto/mock-proto.ts
