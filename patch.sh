sed -i 's/it("opens customization menu", async () => {/it("opens customization menu", async () => { vi.useRealTimers(); /' ui/src/components/dashboard/dashboard-grid.test.tsx
