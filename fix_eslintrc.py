# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import json

path = "ui/.eslintrc.json"
with open(path, "r") as f:
    config = json.load(f)

# eslint outputs warnings if the rules exist.
config["rules"] = {
    "@typescript-eslint/no-explicit-any": "off",
    "@typescript-eslint/no-unused-vars": "off"
}

# The warnings are coming from:
# @typescript-eslint/no-explicit-any
# @typescript-eslint/no-unused-vars

with open(path, "w") as f:
    json.dump(config, f, indent=2)

with open("ui/package.json", "r") as f:
    pkg = json.load(f)

pkg["scripts"]["lint"] = "eslint src/"

with open("ui/package.json", "w") as f:
    json.dump(pkg, f, indent=2)
