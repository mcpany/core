# Market Sync 2026-06-18: Enterprise Traffic Heatmap

## Overview

**Vision:** Elevate the MCP Any network visualizer (`AgentFlow`) into a premium, "Apple-level" observability tool. Currently, the topology mapping is functional but lacks the visceral, immediate impact required by enterprise swarm orchestrators.

**Critical User Journey (CUJ):**
As a Lead Swarm Engineer, when I enable "Live" mode on the Recursive Context Dashboard, I need to instantly, pre-attentively identify cognitive bottlenecks, high-volume routing paths, and overloaded agent mailboxes. The data should flow across the screen with organic fluidity, not mechanical rigidity.

## Aesthetic Spec: The "Pulse" Interface

1. **Color Palette (Heat-Mapped):**
   *   **Idle/Low (0-10 QPS):** `var(--muted)` with a subtle opacity (`0.4`).
   *   **Active (11-100 QPS):** Deep Cobalt Blue (`#3b82f6`) shifting to Amethyst Purple (`#8b5cf6`).
   *   **Saturated (101-500 QPS):** Warm Amber (`#f59e0b`).
   *   **Critical (500+ QPS):** Crimson Red (`#ef4444`) with a localized glow effect (`drop-shadow(0 0 8px rgba(239, 68, 68, 0.6))`).

2. **Animation Physics:**
   *   Edges (`TrafficEdge`) must use a bezier curve with an animated, dash-offset stroke.
   *   The speed of the dash-offset is exponentially proportional to the traffic volume (e.g., `animation-duration: calc(10s / max(1, QPS))`).
   *   Nodes under heavy load should subtly "breathe" (scale transform `1.0` to `1.02` at a frequency tied to error rate or saturation).

3. **Typography & Metrics:**
   *   Use a monospace font (`JetBrains Mono` or similar) for live metric overlays.
   *   Transitions between metric numbers must use a rolling counter animation, eliminating harsh, abrupt jumps when data refreshes.

## Actionable Next Steps

- Refactor `TrafficEdge` to consume a dynamic `heatIndex` property.
- Update node renderers (`AgentNode`, `ToolNode`) to accept load state indicators.
- Seed a "Swarm Load Balancer" or high-volume test server to adequately demonstrate the aesthetic in the UI.
