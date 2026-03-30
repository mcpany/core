# Premium Tool Execution Timeline: Aesthetic RFC

## 1. Overview
The "Premium Tool Execution Timeline" is a high-fidelity, interactive React component designed to visualize the lifecycle of an MCP Any tool execution. It transforms a raw JSON execution trace (comprising pre-hooks, middleware interceptions, upstream network calls, and post-hooks) into an "Apple-level" interactive waterfall diagram.

**Goal:** If a UI component doesn't feel premium, it is a bug. This component must feel incredibly smooth, visually distinct, and instantly readable.

## 2. Customer User Journey (CUJ)
**Persona:** L7 Engineer / Architect debugging a complex tool failure across multiple middleware boundaries.

**The Journey:**
1. The user executes a tool via the `ToolRunner` playground or clicks on a trace from the "Recent Activity" widget.
2. Instead of a flat JSON blob or a simple bulleted list, the user is presented with a vertical, animated timeline.
3. The timeline dynamically unfolds, showing the exact sequence of events: `Pre-Call Hooks -> Upstream Call -> Post-Call Hooks`.
4. The user can clearly see durations (e.g., `45ms`) and status indicators (success, warning, error) for each discrete step.
5. The user can expand a specific node (e.g., a failing middleware) to inspect the exact payload or error message that was mutated or returned.

## 3. Aesthetic Spec & Design Language

### 3.1 Layout & Structure
*   **Vertical Waterfall:** A central vertical axis connects all nodes. Nodes alternate left/right or stack cleanly depending on available width.
*   **Progressive Disclosure:** Nodes are collapsed by default, showing only Title, Duration, and Status Icon. Expanding a node reveals the JSON payload or raw logs.
*   **Visual Hierarchy:** The "Upstream Call" is visually anchored as the primary event, with hooks acting as modifiers before and after.

### 3.2 Colors & Status Indicators
We utilize the existing Tailwind design system but apply it with high contrast and subtle glows.
*   **Success (Passed):** `text-emerald-500`, subtle `bg-emerald-500/10` background for the node.
*   **Warning (Modified/Interrupted):** `text-amber-500`, subtle `bg-amber-500/10` background.
*   **Error (Failed):** `text-rose-500`, subtle `bg-rose-500/10` background.
*   **Timeline Axis:** `border-muted` (subtle gray/slate).
*   **Active/Selected Node:** Stronger border, elevated shadow (`shadow-md`), and a subtle inner ring.

### 3.3 Typography
*   **Titles:** `font-semibold text-sm text-foreground` (e.g., "Authentication Middleware").
*   **Metadata (Duration, Timestamps):** `font-mono text-xs text-muted-foreground` (e.g., `12ms`).
*   **Descriptions:** `text-xs text-muted-foreground`.

### 3.4 Animation & Transitions (Framer Motion / CSS)
*   **Entrance:** Staggered fade-in and slight slide-up (`translate-y-2` to `translate-y-0`) for nodes as they appear.
*   **Expansion:** Smooth height transition when expanding a node to view payload details.
*   **Hover State:** Subtle scale (`scale-[1.01]`) and brightness increase on hover to indicate interactivity.

### 3.5 Icons (Lucide React)
*   **Pre-Hook:** `ArrowRightCircle` or `Filter`
*   **Upstream:** `CloudLightning` or `Server`
*   **Post-Hook:** `CheckCircle2` or `FileOutput`
*   **Error:** `AlertTriangle` or `XCircle`

## 4. Component API Blueprint

```typescript
// ui/src/components/ui/timeline.tsx

export type TimelineStatus = 'success' | 'warning' | 'error' | 'pending';

export interface TimelineEvent {
  id: string;
  title: string;
  description?: string;
  timestamp: string; // ISO string or formatted time
  durationMs?: number;
  status: TimelineStatus;
  icon?: React.ReactNode;
  metadata?: Record<string, any>; // JSON payload to inspect
}

export interface TimelineProps {
  events: TimelineEvent[];
  className?: string;
}
```

## 5. Integration Points
*   **Playground (`ToolRunner`):** Add a "Timeline" tab next to the "Raw Output" and "Rendered" tabs in the `RichResultViewer` or directly inside the execution result panel if tracing data is available.
*   **Traces Dashboard:** Replace standard table rows with this timeline view when inspecting a specific trace detail.
