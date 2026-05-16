# Aesthetic RFC: Agent Cognitive Load Monitor

**Date:** 2026-11-02
**Status:** Accepted
**Author:** Product Architect & Lead Engineer (L7)

## 1. Context & Motivation

As MCP Any expands into complex multi-agent orchestrations, understanding the real-time "cognitive stress" of individual agents becomes critical. Currently, developers have no visceral way to see if an agent is nearing its context window limit or experiencing high latency due to excessive reasoning steps.

The **Agent Cognitive Load Monitor** addresses this friction point. It provides an "Apple-level" premium visualization of an agent's active memory usage, context shard saturation, and real-time cognitive stress. If an agent is overwhelmed, this widget makes it instantly apparent through elegant, hardware-accelerated visual cues rather than relying on abstract metric logs.

## 2. Critical User Journey (CUJ)

1.  **Discovery:** The user navigates to the Agent Profile or Dashboard.
2.  **Engagement:** A new widget, "Cognitive Load Monitor", occupies a `third` or `half` width slot. It displays a minimalist, dynamic set of concentric glowing rings representing different facets of the agent's state (Context Window, Scratchpad, Attention).
3.  **Observation:** The user watches the rings fluctuate as the agent processes intents.
    *   *Healthy State:* The rings are thin, smooth, and pulse gently with a soft `emerald-400` or `cyan-400` gradient.
    *   *High Load:* As the context window fills, the outer ring thickens and transitions to a warm `amber-500`. The pulsing heartbeat quickens.
4.  **Anomaly Detection:** An agent experiences a "Context Smearing" event or hits its token limit. The ring turns a stark `rose-500`, the smooth animation stutters intentionally (simulating strain), and a frosted-glass tooltip drops down: "Warning: Cognitive Overload - Context Pruning Initiated."
5.  **Drill-down:** Clicking the widget opens a modal showing the exact token distribution across recent thoughts, allowing the developer to tune the system prompt.

## 3. Aesthetic Spec & UI Component Design

This widget must feel like a precision instrument—akin to the activity rings on an Apple Watch, but designed for advanced AI operations.

### 3.1. Visual Language & Typography
*   **Theme:** Deep dark mode (`bg-zinc-950`). The rings must provide the primary light source for the component.
*   **Typography:** Inter (or similar geometric sans). Active metrics (e.g., `32K / 128K TOKENS`) use tabular numbers (`font-mono text-[10px] uppercase tracking-widest text-slate-400`).
*   **Depth:** Heavy use of inset shadows and backdrop blur to make the rings feel embedded in a glass display.

### 3.2. Graph Elements
*   **Concentric Rings (Activity Rings):**
    *   **Outer Ring (Context Window):** Represents total token usage. Gradient stroke that shifts color based on saturation (Cyan -> Amber -> Rose).
    *   **Middle Ring (Scratchpad Space):** Represents short-term working memory usage.
    *   **Inner Ring (Attention Sizzle):** Represents active reasoning. A dashed, rotating ring that spins faster when the agent is actively invoking tools.
*   **Center Display:** A crisp, anti-aliased icon (e.g., `Brain` or `Cpu`) surrounded by the current "Stress Score" percentage.

### 3.3. Technical Implementation Details (React/Tailwind)

*   **Container:** `Card` wrapper with a subtle noise texture overlay (`opacity-5`) to give it a matte physical feel.
*   **Animation Library:** Framer Motion for the fluid, physics-based filling of the rings and color interpolation.
*   **Rendering:** SVG with `<circle>` elements using `stroke-dasharray` and `stroke-dashoffset` for the ring progress. CSS filters (`drop-shadow`) for the neon glow effect.

## 4. Implementation Readiness
*   **Frontend:** The widget will be implemented as `ui/src/components/dashboard/cognitive-load-widget.tsx` and registered in `widget-registry.tsx`.
*   **Backend (Seeding):** `server/pkg/app/seeds.go` will include a `cognitive-load-telemetry` service mock that streams realistic token counts, stress scores, and simulated overload events to test the visual fidelity immediately.
*   **Execution:** Actionable immediately. Zero human approval required. Proceeding to seeding.
