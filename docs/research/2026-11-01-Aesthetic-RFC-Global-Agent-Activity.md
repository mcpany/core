# MCP Any - Aesthetic Spec: Global Agent Activity Map

## 1. Goal
Introduce a "Global Agent Activity Map" to the MCP Any dashboard. This visualization will act as a premium, "Apple-level" interactive element that displays real-time (or simulated real-time) geolocated tool usage across the MCP ecosystem.

## 2. Friction Point & Justification
Currently, the dashboard lacks a highly visual, spatial representation of where and how agents are operating globally. While lists and charts are functional, a dynamic, glowing map showing agent activity (like nodes lighting up or arcs connecting regions) provides an immediate, visceral understanding of scale and operation. A premium developer experience demands visualization that feels alive and impactful.

## 3. Aesthetic Spec

### Colors & Typography
*   **Map Background:** Deep, subtle dark mode gradient or solid dark background (e.g., `#0f172a` or `bg-slate-900`) to make the glowing nodes pop.
*   **Nodes (Agents/Tools):**
    *   **Active:** Glowing neon cyan (`#06b6d4` or `text-cyan-500`) with a soft drop shadow (`drop-shadow-[0_0_8px_rgba(6,182,212,0.8)]`).
    *   **Idle/Background:** Muted slate (`#64748b` or `text-slate-500`).
    *   **Error/Alert:** Vibrant crimson (`#e11d48` or `text-rose-600`) with a pulsing effect.
*   **Connections (Arcs/Lines):** Thin, semi-transparent curved lines connecting active nodes, using a gradient from cyan to indigo.
*   **Typography:** Strict use of Inter (or system sans-serif) for tooltips and labels, ensuring high contrast against the dark map background.

### Interactions & Animations
*   **Pulsing Nodes:** Active locations should have a subtle, continuous radial pulse animation to indicate live activity.
*   **Hover States:** Hovering over a node should smoothly expand (`scale-110`) the node and reveal a frosted-glass tooltip (`backdrop-blur-md bg-background/80`) detailing the specific agent, tool executed, and latency.
*   **Transitions:** The map initialization and node appearances should fade in smoothly (`transition-opacity duration-700 ease-out`).

### Core Components
*   **The Globe/Map Canvas:** A responsive SVG or WebGL-based (if available, otherwise high-quality SVG) map of the world.
*   **Activity Tooltip:** A sleek, minimal popover showing live telemetry data for a specific geographic node.

## 4. Implementation Readiness
*   **Component Creation:** Create `ui/src/components/dashboard/global-activity-map-widget.tsx` to implement this visual.
*   **Seeding Update:** Enhance `server/pkg/app/seeds.go` to include a `global-telemetry` service or update existing mocks to provide rich, geolocated demonstration data (e.g., coordinates, agent names, tool usage stats) to power this visualization instantly upon deployment.
