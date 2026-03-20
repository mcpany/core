# Aesthetic RFC: "Live MCP Data Flow Visualizer"

**Date:** 2026-06-25
**Author:** Product Architect & Lead Engineer
**Status:** Approved for Implementation
**Target Release:** Premium "Portainer for MCP" Update

## 1. Context and Problem Statement
MCP Any currently provides static topological views of networks and services. However, a key element of the "Apple-level" Portainer-for-MCP experience is feeling the *pulse* of the system. Users lack a visceral, immediate understanding of real-time data flow, reasoning execution, and shard locking as they happen across their multi-agent setups. We need a premium, high-fidelity visualizer that transitions us from "configuration viewing" to "live telemetry immersion."

## 2. Core Customer User Journey (CUJ)
**The "God Mode" Observer Journey:**
1. A user (Swarm Commander) navigates to the "Live Data Flow" or clicks a "Live View" button on their dashboard.
2. The UI instantly transitions into a dark-mode, high-contrast God-Mode view.
3. Services, Tools, and Agents are represented as premium glowing nodes.
4. When an AI client makes a request to a tool, a high-frame-rate particle stream or glowing pulse (a "Data Packet") visually travels along the bezier curves connecting the nodes.
5. If an error occurs, the packet glows crimson, shatters, or creates a ripple effect, instantly drawing the eye to the point of failure.
6. The user can hover over any live stream to pause the flow and inspect the payload/reasoning context in a beautifully frosted-glass modal.

## 3. Aesthetic Spec & Design Language
### 3.1. Visual Theme (The "Steve Jobs" Touch)
*   **Base:** Deep Obsidian (#09090B) to Jet Black (#000000). The background must recede entirely.
*   **Typography:** Inter for utility, Geist Sans for headers. Weights must be strictly regulated (Medium for labels, Semibold for active states).
*   **Materials:** Frosted glass (backdrop-filter: blur(12px)) with ultra-thin, low-opacity borders (rgba(255,255,255,0.08)) for floating panels.

### 3.2. Color Palette & States
*   **Idle / Healthy Node:** Soft Cobalt Pulse (#3B82F6) to Cyan (#06B6D4).
*   **Active Processing:** Neon Emerald (#10B981) with a 0px 0px 15px glowing drop-shadow.
*   **Blocked / Error:** Crimson Red (#EF4444) pulsing aggressively.
*   **Data Packets (The Traffic):** Brilliant White (#FFFFFF) with a trail of fading primary color (Cyan/Emerald based on context).

### 3.3. Animations & Transitions
*   **Frictionless:** Zero jank. We target stable 60/120fps using hardware-accelerated CSS or Canvas/WebGL.
*   **Packet Movement:** Easing functions must be silky. Custom cubic-bezier(0.25, 1, 0.5, 1) to give weight and momentum to the data.
*   **Node Interaction:** Hovering over a node triggers a micro-interaction: a 1.05x scale bounce and an intensification of its outer glow over 150ms.

### 4. Implementation Requirements
*   **Engine:** Leverage `@xyflow/react` but strip away its default styles. Inject our own custom edge rendering.
*   **Custom Edge:** Implement a `<LiveTrafficEdge />` component that renders an SVG path and animates an SVG `<circle>` or `<rect>` along that path using `stroke-dasharray` and CSS animations.
*   **Data Source:** The backend must provide a stream (via WebSocket or frequent polling) of recent tool executions to feed the visualizer.

## 5. Exit Criteria
- The visualizer feels objectively "expensive."
- Zero lag during normal operation.
- No human design approval required; this document is the absolute source of truth.
