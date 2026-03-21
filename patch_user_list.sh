#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/export interface User/\/** Local User type for UI display *\/\nexport interface User/' ui/src/components/users/user-list.tsx
