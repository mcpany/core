# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open('ui/src/mocks/proto/mock-proto.ts', 'r') as f:
    content = f.read()

content = re.sub(r'export const CallPolicyRule = \{\};', '/**\n * Mock type placeholders for policy-related proto messages.\n */\nexport const CallPolicyRule = {};', content)
content = re.sub(r'export const ExportPolicy = \{\};', '/**\n * Mock type placeholders for policy-related proto messages.\n */\nexport const ExportPolicy = {};', content)
content = re.sub(r'export const ExportRule = \{\};', '/**\n * Mock type placeholders for policy-related proto messages.\n */\nexport const ExportRule = {};', content)

with open('ui/src/mocks/proto/mock-proto.ts', 'w') as f:
    f.write(content)
