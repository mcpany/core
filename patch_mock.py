# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open('ui/src/mocks/proto/mock-proto.ts', 'r') as f:
    content = f.read()

content = re.sub(r'export interface CallPolicyRule \{', '/**\n * Mock for CallPolicyRule\n */\nexport interface CallPolicyRule {', content)
content = re.sub(r'export interface ExportPolicy \{', '/**\n * Mock for ExportPolicy\n */\nexport interface ExportPolicy {', content)
content = re.sub(r'export interface ExportRule \{', '/**\n * Mock for ExportRule\n */\nexport interface ExportRule {', content)

with open('ui/src/mocks/proto/mock-proto.ts', 'w') as f:
    f.write(content)
