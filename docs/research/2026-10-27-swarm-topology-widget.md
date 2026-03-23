# Aesthetic RFC: Multi-Agent Swarm Topology & Sovereignty Monitor

**Date:** 2026-10-27
**Status:** Accepted
**Author:** Product Architect & Lead Engineer (L7)

## 1. Context & Motivation

As MCP Any evolves into the "Portainer for MCP," managing the complexity of multi-agent interactions becomes paramount. Current dashboards excel at individual service metrics, but fail to convey the dynamic, ephemeral, and cryptographically entangled nature of a multi-agent "swarm."

The **Multi-Agent Swarm Topology & Sovereignty Monitor** addresses the P0 requirements outlined in our strategic roadmap (e.g., "Swarm Anomaly Visualizer", "Mission-Locked Execution (MLE) Visualizer", "Teammate Handshake Monitor"). It provides an "Apple-level" premium visualization of real-time agent coordination, context sharding, and cryptographic sovereignty.

If an agent attempts an unauthorized state splice or experiences high cognitive latency (MSHE), this widget makes it immediately obvious through visceral, hardware-accelerated animations.

## 2. Critical User Journey (CUJ)

1.  **Discovery:** The user navigates to the main Dashboard.
2.  **Engagement:** A new widget, "Swarm Topology," occupies a `full` or `two-thirds` width slot by default. It displays a dynamic, physics-based force-directed graph (or an elegant radial equivalent) of active agent nodes.
3.  **Observation:** The user watches "Intent Fragments" (data packets) flow between nodes.
    *   *Healthy paths:* Rendered as smooth, glowing splines (Tailwind `emerald-500` to `cyan-400` gradients).
    *   *Sovereignty Checks:* Brief, cryptographic "locking" animations occur at nodes during HAAL (Hardware-Attested Attention Locking).
4.  **Anomaly Detection:** An agent attempts an unauthorized logic graft (Semantic Drift). The node pulses `destructive` (red), the connection is visually severed with a shattered-glass/static effect, and a toast notification ("ARI Hub: Logic Graft Blocked") appears.
5.  **Drill-down:** Clicking a node reveals its "Ephemeral Mission Root" and active capability leases in a sleek, glassmorphic side-panel.

## 3. Aesthetic Spec & UI Component Design

This widget must not look like a generic D3 chart. It must feel like a mission-control HUD built by a luxury design firm.

### 3.1. Visual Language & Typography
*   **Theme:** Dark mode biased, with high-contrast, glowing accents against a deep, muted background (`bg-slate-950` or `bg-zinc-950`).
*   **Typography:** Inter or a similarly geometric sans-serif. Use tabular nums for any active metrics (e.g., `font-mono text-[10px] uppercase tracking-widest`).
*   **Depth:** Heavy use of `backdrop-blur` (glassmorphism) for overlays and tooltips, supported by subtle inner borders (`border-white/10`).

### 3.2. Graph Elements
*   **Nodes (Agents/Services):**
    *   Shape: Perfect circles or softly rounded squarcles (`rounded-2xl`).
    *   Idle State: Subtle pulsating glow indicating "Active Intent Alignment" heartbeat.
    *   Active State (Executing): A rotating dashed border or an inner conic-gradient spin.
    *   Icons: Lucide icons (e.g., `Bot`, `Server`, `Shield`) centered within the node, perfectly anti-aliased.
*   **Edges (Connections/Handshakes):**
    *   Style: Curved SVG paths (Bézier curves), not straight lines, to feel organic.
    *   Flow: Animated SVG `stroke-dashoffset` or moving particles along the path to simulate "Context Fragments" streaming.
    *   Hardware-Locked (Secure): Solid, vibrant cyan/emerald gradients.
    *   Speculative/Unverified: Dashed, lower opacity (`opacity-40`), perhaps a subtle amber tint.
*   **Anomalies (The "Sizzle"):**
    *   When an intent-splicing attempt is blocked, the edge should visually "snap" or dissolve into a red gradient with a localized particle explosion effect (using Framer Motion or lightweight CSS animations).
    *   Nodes experiencing "Cognitive Stall" vibrate slightly or pulse a warning color (`amber-500`).

### 3.3. Technical Implementation Details (React/Tailwind)

*   **Container:** Uses a `Card` wrapper but with a deeply inset background map or grid pattern (e.g., an SVG dot grid with radial fade `[mask-image:radial-gradient(ellipse_at_center,transparent_20%,black)]`).
*   **Animation Library:** Framer Motion for node entry/exit and layout transitions.
*   **Canvas vs. SVG:** For `< 50` nodes (typical swarm), SVG is perfectly performant and easier to style with Tailwind. For larger scales, a WebGL layer (like React Flow or a custom canvas) might be needed, but we will target SVG for the premium crispness of CSS filters.

## 4. Interaction Design

*   **Hover:** Hovering over a node dims the rest of the graph (`opacity-20`), highlighting only its direct upstream/downstream dependencies (the "Mesh Command Sovereignty" chain).
*   **Click:** Selecting an edge opens a popover detailing the "Teammate Handshake" status (TPM Attestation, Spectral Jitter ms, Fragment Sovereignty).

## 5. Implementation Readiness
*   **Frontend:** The widget will be implemented in `src/components/dashboard/swarm-topology-widget.tsx`. It will register via `widget-registry.tsx`.
*   **Backend (Seeding):** `server/pkg/app/seeds.go` will be updated to include a mock "Swarm Orchestrator" service that generates synthetic topology data (nodes, edges, and simulated anomalies like CVE-2026-62001 attempts).
*   **Execution:** Actionable immediately. Zero human approval required. Proceeding to implementation.
