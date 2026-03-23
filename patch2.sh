# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/it("opens customization menu", async () => { vi.useFakeTimers();  /it("opens customization menu", async () => { vi.useRealTimers(); /' ui/src/components/dashboard/dashboard-grid.test.tsx
