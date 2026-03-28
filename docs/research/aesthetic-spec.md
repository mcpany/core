# Aesthetic Spec: Premium "Portainer for MCP"

## 1. Overview
The goal is to elevate the MCP Any dashboard from a functional UI to an "Apple-level" premium experience. We are targeting the "Recent Activity" widget, which currently lacks visual polish and feels like a raw data feed rather than a sleek timeline of events.

## 2. Core Principles
- **Glassmorphism & Depth:** Soft drop shadows, translucent backgrounds, and subtle borders to create a layered, spatial feel.
- **Micro-Interactions:** Smooth, physics-based transitions for hover states, skeleton loading, and list expansion.
- **Typographic Hierarchy:** High-contrast primary text with muted, legible secondary text using a modern sans-serif font stack.
- **Color Palette:**
  - Backgrounds: `bg-white/50 dark:bg-zinc-900/50` with backdrop blur.
  - Borders: `border-white/20 dark:border-zinc-800/50`.
  - Accents: iOS-style system blue (`#007AFF`) and system green (`#34C759`) for status indicators.

## 3. Targeted Component: Recent Activity Widget
### Current Friction Point
The current activity feed is visually flat. It lacks a clear distinction between different event types, making it hard to parse quickly.

### The "Gold Standard" Implementation
1.  **Timeline Visuals:** Replace basic list items with a continuous timeline track.
2.  **Iconography:** Use soft, rounded-rect icons with varied background tints based on the event type (e.g., green for success, blue for tool calls, purple for system events).
3.  **Inline Details:** When expanded, show a premium inline diff or JSON tree for complex payload data (e.g., trace bodies).
4.  **Animations:** Use `framer-motion` (or standard CSS transitions) to slide and fade in new items.

## 4. Seeder Requirements
To demonstrate this premium feel, the mock data (seeders) must be robust.
- Provide varied tool calls (e.g., "code-refactor", "database-query").
- Include realistic durations (e.g., 120ms, 45ms).
- Status codes indicating success (200) and occasional warnings or errors to show off the styling states.
