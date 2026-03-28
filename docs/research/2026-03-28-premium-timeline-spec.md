# MCP Any - Aesthetic Spec: Premium Tool Execution Timeline

## 1. Goal
Elevate the "Tool Execution History" view (often called "Recent Activity") from a basic, utilitarian list into a premium, "Apple-level" interactive timeline. This feature allows users to inspect, replay, and debug tool executions with a sleek, high-fidelity interface.

## 2. Friction Point & Justification
Currently, visualizing what an agent or user just executed feels disjointed and lacks polish. The logs are functional but not beautiful. A premium developer experience demands a unified, animated, and clear timeline of events (tool calls) that feels instantly responsive and visually hierarchical. If it doesn't feel premium, it's a bug.

## 3. Aesthetic Spec

### Colors & Typography
*   **Primary Background:** Deep, subtle blur (Glassmorphism) using `bg-background/80 backdrop-blur-md` for the main container.
*   **Success State:** A soft, glowing emerald `text-emerald-500` with a subtle background tint `bg-emerald-500/10`.
*   **Error State:** A vibrant, clear destructive red `text-destructive` with `bg-destructive/10`.
*   **Typography:** Strict use of Inter (or system sans-serif) with clear weight hierarchy. Mono-spaced fonts (e.g., Fira Code, JetBrains Mono) strictly reserved for code blocks and JSON payloads.

### Interactions & Animations
*   **Transitions:** All state changes (expanding a tool call payload, hovering over a row) must use smooth, spring-like transitions (`transition-all duration-300 ease-in-out`).
*   **Hover States:** Subtle scale-up (`scale-[1.01]`) and shadow elevation (`shadow-md`) on interactive timeline cards.

### Core Components
*   **The Timeline Node:** A custom dot indicator with a pulsating ring for active/recent executions.
*   **The Payload Diff:** When a tool is re-run and the output changes, a unified inline diff (red/green background highlights) is shown, styled like a premium code editor (e.g., VS Code or Linear's diff viewer).
*   **Rich Result Viewer:** Structured data payloads should be seamlessly rendered via a RichResultViewer to make large blocks of data readable and expandable rather than raw text blobs.

## 4. Implementation Readiness
*   **Component Update:** Modified `ui/src/components/dashboard/recent-activity-widget.tsx` to implement this timeline aesthetic using `RichResultViewer`.
*   **Seeding Update:** Enhancing `server/pkg/app/seed.go` and `server/pkg/app/seeds.go` to ensure the timeline has rich, realistic "Gold Standard" mock data (e.g., long JSON responses, arrays of objects, varied schemas) for immediate visual verification.
