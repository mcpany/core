# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

sed -i 's/it("opens customization menu", async () => {/it("opens customization menu", async () => { vi.useFakeTimers({ shouldAdvanceTime: true }); /' ui/src/components/dashboard/dashboard-grid.test.tsx
