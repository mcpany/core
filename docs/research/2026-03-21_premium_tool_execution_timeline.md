# Design Doc: Premium Tool Execution Timeline

## Abstract
The Universal Agent Bus requires deep observability into agent handoffs, latency, and tool execution boundaries. The current logging mechanisms provide necessary data but lack the "Apple-level" interactive polish needed for an L7 product. This document specifies the Critical User Journey (CUJ) and Aesthetic Specification for the Premium Tool Execution Timeline, which replaces the raw trace tables with a high-fidelity, interactive sequence viewer.

## Critical User Journey (CUJ)
1. **Trigger**: An agent executes a complex sequence of tool calls across multiple swarms.
2. **View**: The user navigates to the "Traces" or "Activity" dashboard.
3. **Interact**:
    - The user scrolls through a smooth, vertical or horizontal timeline.
    - Hovering over a tool execution expands a glassmorphic popover showing exact request/response payloads and token consumption.
    - Expanding nested subagent handoffs reveals recursive execution layers with staggered animations.
4. **Outcome**: The user instantly identifies latency bottlenecks, failed tool paths, and token consumption boundaries without deciphering JSON logs.

## Aesthetic Specification
The timeline must adhere to our premium standard, feeling tactile and deeply responsive.

### Colors
- **Background**: Translucent/glassmorphism (e.g., `rgba(18, 18, 20, 0.6)` for dark mode, `rgba(250, 250, 250, 0.7)` for light mode) with backdrop-filter blur (`blur(16px)`).
- **Accents**: Neon, glowing accents for execution states:
    - *Success*: Emerald green (`#10b981`) with a subtle `box-shadow` glow.
    - *In Progress/Pending*: Electric blue (`#3b82f6`) with a pulsing animation.
    - *Error/Interdicted*: Crimson red (`#ef4444`) with a sharp, immediate presence.
    - *Agent Node*: Violet (`#8b5cf6`) denoting handoffs to specialized models.
- **Lines/Connectors**: Subtle gradient lines connecting nodes, fading to transparent to indicate ongoing processes.

### Fonts
- **Primary Font**: `Inter` or `SF Pro Display` (system-default premium sans-serif).
- **Weights**:
    - Headers: `600` (SemiBold) for tool names and agent IDs.
    - Metadata/Timestamps: `400` (Regular) for latency and payload sizes.
    - Monospace (Code/Payloads): `JetBrains Mono` or `Fira Code` for JSON/text snippets.

### Transitions & Animations
- **Physics**: Spring physics (e.g., Framer Motion's `spring` config with `stiffness: 300`, `damping: 30`) for all node expansions, popovers, and hover states. No linear easing.
- **Stagger**: When loading a sequence of nodes, they must cascade into view with a stagger effect (`staggerChildren: 0.05`).
- **Hover Reveal**: Payloads and deep metrics reveal themselves with a smooth opacity fade-in (`duration: 0.2s`) and slight vertical slide-up (`translateY: 4px` to `0`).

## Execution Logic & Seeding
To support this UI component, a robust backend seeder is required to generate mock traces that mimic high-fidelity telemetry.

**Mock Telemetry Provider**: `premium-timeline-tracer`
- **Capabilities**: Simulates deeply nested tool executions, intentional latency variations, and cross-framework agent handoffs.
- **Attributes**: Will emit JSON schemas matching our `Trace` definition, complete with timestamps, token counts, and mock payload diffs.

## Exit Criteria
- Timeline component passes all visual regression tests.
- Seeder successfully injects mock traces into the development environment.
- Zero jank (maintains 60fps) during node expansion with deeply nested execution graphs.