# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open("server/pkg/app/server_test.go", "r") as f:
    content = f.read()

new_content = content.replace("app := NewApplication()", "app := NewApplication()\n\tapp.testMode = true")
with open("server/pkg/app/server_test.go", "w") as f:
    f.write(new_content)
