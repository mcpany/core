# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re
# Original: greedy
# FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\((.*)\)\s*(.*)\{')
# New: non-greedy
FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*)\{')

line = "func NewWatcher() (*Watcher, error) {"
m = FUNC_PATTERN.match(line)
print(f"Match: {m}")
if m:
    print(f"Group 1 (Name): '{m.group(1)}'")
    print(f"Group 2 (Params): '{m.group(2)}'")
    print(f"Group 3 (Returns): '{m.group(3)}'")
