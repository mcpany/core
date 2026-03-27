# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import json

def list_missing():
    # just run the tool
    pass

with open("missing_files.json", "r") as f:
    files = json.load(f)

print(f"Total files updated: {len(files)}")
