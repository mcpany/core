# MCP Any - Portainer for MCP Aesthetic Spec

## 1. Problem Statement
The current default dashboard simply dumps all available widgets onto the screen using `WIDGET_DEFINITIONS.map(...)`. This leads to a cluttered, unopinionated, and overwhelming out-of-the-box experience. A premium "Apple-level" product should present a curated, visually balanced default layout that guides the user's attention to the most critical system vitals (Metrics, Topology, Health) while tucking secondary information away.

## 2. Solution: Curated Default Dashboard Experience
We will modify the `DEFAULT_LAYOUT` in `ui/src/components/dashboard/dashboard-grid.tsx` to present a clean, organized default view.

**Target Default Layout:**
1.  **Row 1:** Metrics Overview (full width) - Essential top-level stats.
2.  **Row 2:** Network Topology (two-thirds width) + Service Health (third width) - Core architectural visualization paired with status.
3.  **Row 3:** Recent Activity (half width) + Request Volume (half width) - Operational insights.
4.  *(Hidden by default: Tool Failure Rates, Quick Actions, System Uptime, Top Tools, Audit Logs, Swarm Topology, Active Intent Alignment Monitor)*

## 3. Seed Data (Gold Standard)
To properly demonstrate this premium layout, the initial seed data must be rich.
We will update `server/pkg/app/seeds_collections.go` to include a "Gold Standard" collection that uses popular/visual tools (e.g., memory, slack, github, sqlite-db) to ensure the Network Topology and Metrics widgets look impressive out of the box.

## 4. Aesthetics & Typography
- **Colors & Fonts:** Stick to the existing Tailwind theme but ensure the curated layout respects the grid cleanly. The combination of full, two-thirds, and third sizes creates a visually appealing masonry-like grid.
- **Transitions:** Use smooth drag-and-drop (already provided by `@hello-pangea/dnd`) but the initial load will feel much lighter and faster due to fewer widgets rendering simultaneously.
