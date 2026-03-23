# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/vi.useFakeTimers({ shouldAdvanceTime: true });/vi.useFakeTimers();/' ui/src/components/dashboard/dashboard-grid.test.tsx
