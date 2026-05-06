# Aesthetic RFC: Global Agent Activity Map

## 1. CUJ (Critical User Journey)
As an MCP Any administrator overseeing a multi-regional agent mesh, I want to see a live, glowing 3D globe visualization of all active agent nodes, their current execution state, and cross-region tool invocation traffic. This ensures I can visually confirm the health, security, and distribution of my global agent network at a glance.

## 2. Visual Spec & "Apple-Level" Aesthetics
**Overall Vibe:** Sleek, dark-mode-first, cinematic. Think "Portainer meets a premium cyber-command center."

*   **Colors:**
    *   Globe Base: Deep Obsidian (`#0B0C10`) with subtle specular highlights.
    *   Active Nodes (Healthy): Premium Cyan (`#00E5FF`) with a soft Gaussian blur for a glowing effect.
    *   Active Nodes (Warning/High Load): Amber (`#FFB300`).
    *   Traffic Arcs: Semi-transparent electric violet (`rgba(138, 43, 226, 0.6)`) to signify data/intent passing between regions.
    *   Background: Radial gradient from `#1F2833` to `#0B0C10`.
*   **Typography:**
    *   SF Pro Display (Apple ecosystem) or Inter for clean, legible metrics overlays.
    *   Monospaced fonts (SF Mono, JetBrains Mono) for specific node coordinates and latency data (e.g., `45ms`).
*   **Transitions & Motion:**
    *   Smooth, continuous, slow rotation of the globe (approx 1 RPM).
    *   Traffic arcs draw themselves using a bezier curve animation over 1.5 seconds, with a leading "spark."
    *   Hovering over a node uses a spring-physics scaling effect (`scale 1.1`, damping 20, stiffness 300) to reveal a sleek glassmorphic tooltip.

## 3. Implementation Requirements
*   **Component:** `GlobalActivityMapWidget`
*   **Data Source:** `/api/v1/telemetry/global` (Seeded by the `Global Telemetry` mock template).
*   **Libraries:** `react-globe.gl` or `three.js` for premium performance.
