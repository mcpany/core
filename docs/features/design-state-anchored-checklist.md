# Design Doc: State-Anchored Checklist Middleware
**Status:** Draft
**Created:** 2026-03-29

## 1. Context and Scope
As AI agents execute complex, multi-step plans, they frequently undergo "Context Compaction" to stay within LLM token limits. During this process, high-level mission metadata (e.g., the current checklist or "Plan Mode" status) is often lost, leading to "Plan Ghosting" where the agent loses its strategic orientation. This middleware provides a "State Anchor" to persist and re-inject critical plan state.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a `ChecklistAnchor` tool that allows agents to store their high-level plan externally.
    * Automatically re-inject the "Active Checklist" into the agent's context following a compaction event.
    * Provide a "Mission North Star" header that is present in every tool call, regardless of context depth.
* **Non-Goals:**
    * Managing the agent's internal reasoning (the agent still decides when to update the checklist).
    * replacing the primary context (it only anchors the *strategic* layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Complex Project Migrator
* **Primary Goal:** Ensure the agent completes a 20-step migration without forgetting the high-level goals after step 5 (when compaction usually occurs).
* **The Happy Path (Tasks):**
    1. Agent initializes the `ChecklistAnchor` with a 20-step plan.
    2. Agent completes 4 steps, bloating its context.
    3. The LLM or Gateway triggers context compaction.
    4. Checklist Middleware detects the compaction and re-fetches the plan from the Shared KV Store.
    5. Middleware re-injects the plan as a "System-Level Anchor" in the next message.
    6. Agent resumes step 5 with full awareness of the remaining 15 steps.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Update Checklist` -> `Shared KV Store (Blackboard)` -> `Compaction Event` -> `Middleware Interceptor` -> `Plan Re-injection` -> `Agent`
* **APIs / Interfaces:**
    * `ChecklistManager`: `SetPlan(steps []Step)`, `GetActivePlan()`, `UpdateStep(id string, status string)`
    * `CompactionHook`: A middleware hook that triggers on context truncation.
* **Data Storage/State:**
    * Plans are stored in the "Plan-Scoped" isolation of the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Manual Re-injection**: Rejected because agents often forget to re-read the plan file after compaction.
* **Context Pinning**: Rejected as it consumes valuable token space that might be needed for the immediate task; the Anchor only re-injects the *summary*.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Checklist updates must be signed by the parent agent to prevent "Plan Hijacking" by subagents.
* **Observability:** The "Mission North Star" is visualized in the "Swarm Rollback Dashboard."

## 7. Evolutionary Changelog
* **2026-03-29:** Initial Document Creation.
