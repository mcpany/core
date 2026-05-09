# Aesthetic RFC: Agent Chain Tracer (A2A) Timeline
Date: 2026-03-01

## 1. Context & CUJ
**Core User Journey (CUJ):**
As a platform operator or agent developer, I need to visualize multi-agent task handoffs (A2A) with cryptographic provenance, reasoning intensity (ARE headers), and latency. The current UI is a flat list of cards. The new UI must be an interactive, "Apple-level" timeline that feels native, performant, and premium.

## 2. Aesthetic Spec & Visual Language
We are adopting a "Glass & Light" aesthetic inspired by premium macOS native applications (e.g., Instruments, Xcode tracing).

- **Colors:**
  - Backgrounds: Translucent glass surfaces (`bg-background/80 backdrop-blur-md`).
  - Accents: Subtle gradients (e.g., `from-blue-500/20 to-purple-500/20` for active reasoning branches).
  - Status Indicators: Neon-esque, high-contrast badges (Emerald for Attested, Amber for Speculative).
- **Typography:**
  - Font: Inter (sans-serif) with tight tracking (`tracking-tight`) for headers.
  - Monospace: JetBrains Mono or UI monospace for cryptographic hashes and token counts.
- **Animations (Framer Motion / CSS):**
  - Hover states: Spring-based scaling (`hover:scale-[1.02]`) and subtle elevation shadows.
  - Reveal: Staggered fade-up for timeline events.
  - Pulse: Slow breathing animations for "Active" agents in the swarm.
- **Layout:**
  - A vertical timeline spine with branching nodes to represent subagent spawns (TeammateTool).
  - Cards expanding on click to reveal cryptographic lineage and raw context diffs.

## 3. Implementation Blueprint
1. **Seeder Update:** Update `BuiltinServiceTemplates` in `server/pkg/app/seeds.go` to include a mock A2A capability or ensure the existing `swarm-orchestrator` has rich tags.
2. **Component:** Create `ui/src/app/universal-agent-bus/agent-chain-tracer.tsx`.
3. **Integration:** Replace the static "Agent Chain Tracer" card in `ui/src/app/universal-agent-bus/page.tsx` with the live, interactive component.
4. **Data:** The component will use mocked React state initially, representing "Gold Standard" A2A handoff data (Agent A -> Agent B with payloads and verification signatures), matching the "Zero Human Opinion" constraint.

## 4. Exit Criteria
- The component must render flawlessly without layout shift.
- Hovering over a timeline node must display a tooltip or expand the card with "Apple-like" fluid transition.
- The design must scream "Premium Enterprise Software."