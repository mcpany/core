# Active Intent Alignment Monitor: Aesthetic & UX Blueprint

## 1. Executive Summary
As part of the Strategic Evolution towards "Mission-Root Sovereignty", the **Active Intent Alignment (AIA) Monitor** is a mandatory P0 security visualization. It provides a real-time, "Apple-level" premium dashboard widget that tracks AIA heartbeats and visualizes "Semantic Drift" across federated subagent swarms.

## 2. Core User Journey (CUJ)
- **Context:** An AI System Administrator is monitoring a long-running, autonomous multi-agent swarm connected via MCP Any.
- **Trigger:** A subagent begins to experience "Reasoning Hijacking" or drifts from the mission-root constraints.
- **Action:** The AIA Monitor instantly highlights the drifting agent with a pulsing, premium visual indicator (moving from a healthy green/blue to a warning amber/red). The administrator can expand the widget to see the exact semantic entropy score and the failing AIA heartbeats.
- **Resolution:** The administrator uses a quick action within the widget to forcefully terminate the drifting sub-session, restoring mission integrity.

## 3. Aesthetic Spec & Vibe
- **Vibe:** "Steve Jobs meets Cyber Command." Unapologetically premium, stark, and data-dense but uncluttered.
- **Color Palette:**
  - **Healthy (Aligned):** Deep, glowing Cyan (`#06b6d4`) or Emerald (`#10b981`) against an ultra-dark background (`#09090b`).
  - **Warning (Drifting):** A sharp, urgent Amber (`#f59e0b`) or Crimson (`#ef4444`).
- **Typography:**
  - San Francisco / Inter for structural elements.
  - Monospace (JetBrains Mono) for hash-chained identity fragments and entropy scores.
- **Motion & Transitions:**
  - **Heartbeat:** A subtle, ease-in-out radial pulse radiating from the center of each agent node to signify a successful AIA heartbeat.
  - **Drift Alert:** A smooth color interpolation accompanied by a subtle, high-frequency "glitch" or shake effect to draw the eye immediately to the anomaly.
  - **State Changes:** CSS-based `backdrop-filter: blur(8px)` with smooth opacity transitions for hover states and modal expansions.

## 4. Implementation Details
- **Location:** `ui/src/components/dashboard/active-intent-alignment-widget.tsx`
- **Integration:** Registered in `WIDGET_DEFINITIONS` within `ui/src/components/dashboard/widget-registry.tsx`.
- **Data Source:** Simulated real-time WebSocket/Polling hook pulling from a new mock endpoint or seeder logic representing AIA status.

## 5. Success Criteria
- The widget feels like a native, premium dashboard component.
- The visual difference between "Aligned" and "Drifting" is immediate and unmistakable.
- Zero human explanation is required to understand the health of the swarm's intent.
