# MCP Any - Aesthetic Spec: Blueprint for Premium Tool Execution Timeline

**Date:** 2026-03-27

## 1. Goal

Elevate the "Tool Execution History" view from a basic, utilitarian list into a premium, "Apple-level" interactive timeline. This feature allows users to inspect, replay, and debug tool executions with a sleek, high-fidelity interface, moving MCP Any towards a sleek, premium "Portainer for MCP" experience.

## 2. Friction Point & Justification

Currently, visualizing what an agent or user just executed feels disjointed and lacks polish. The logs are functional but not beautiful. A premium developer experience demands a unified, animated, and clear timeline of events (tool calls) that feels instantly responsive and visually hierarchical. If it doesn't feel premium, it's a bug.

## 3. Aesthetic Spec

### Colors & Typography
*   **Primary Background:** Deep, subtle blur (Glassmorphism) using `bg-background/80 backdrop-blur-md` for the main container. Use a subtle `border-border/40` on the card to ground it.
*   **Success State:** A soft, glowing emerald `text-emerald-500` with a subtle background tint `bg-emerald-500/10` and `border-emerald-500/20` around nodes.
*   **Error State:** A vibrant, clear destructive red `text-destructive` with `bg-destructive/10` and `border-destructive/20`.
*   **Typography:** Strict use of Inter (or system sans-serif) with clear weight hierarchy. Mono-spaced fonts (e.g., Fira Code, JetBrains Mono) strictly reserved for code blocks and JSON payloads.
*   **Card Backgrounds:** For expanded state payloads, use a deeply recessed `#1e1e1e` (dark) or `#0d0d0d` inner shadow container to represent "raw data".

### Interactions & Animations
*   **Transitions:** All state changes (expanding a tool call payload, hovering over a row) must use smooth, spring-like transitions (`transition-all duration-300 ease-in-out`).
*   **Hover States:** Subtle scale-up (`scale-[1.01]`) and shadow elevation (`shadow-md`) on interactive timeline cards. Cards should layer above siblings (`z-20`) when hovered.

### Core Components
*   **The Timeline Node:** A custom dot indicator with a pulsating ring for active/recent executions. Use a solid `h-5 w-5` rounded full node with a 3px border to match the background, creating a cutout effect on the connector line.
*   **The Connector Line:** A vertical rule `w-px bg-border/50` tying the events together sequentially.
*   **The Payload Diff:** When a tool is re-run and the output changes, a unified inline diff (red/green background highlights) is shown, styled like a premium code editor (e.g., VS Code or Linear's diff viewer). The diff must include left-borders (`border-l-2`) colored to match the addition/removal state.

## 4. Seeding Spec (Gold Standard Data)
The system must be seeded with data that allows immediate validation of this CUJ:
*   A recent successful run.
*   A failed run with a lengthy timeout error message.
*   A run returning a lengthy, nested JSON payload.
*   A run returning a `mcp.response_diff` that demonstrates additions, removals, and context lines for the diff viewer component.
