# Design Doc: Narrative Continuity Controller (NCC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The launch of Gemini CLI subagents and Claude Code's "Dispatch" mode has introduced a "Narrative Fragmentation" problem. As tasks are delegated to specialized experts across different frameworks, the high-level mission story ("Narrative") is lost in isolated context windows. MCP Any needs to act as the universal bridge that synchronizes these narrative "Chapters" and state fragments to ensure long-term mission coherence across heterogeneous agent teams.

## 2. Goals & Non-Goals
* **Goals:**
    * Synchronize narrative "Chapters" across Gemini subagents and Claude Dispatch workers.
    * Provide a unified state view for long-running agent missions.
    * Enable seamless context inheritance between disparate framework workers.
* **Non-Goals:**
    * Replacing framework-specific memory management.
    * Managing the internal reasoning loop of the agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Maintain a consistent narrative story across 5 Gemini subagents and 3 Claude background workers during a complex migration.
* **The Happy Path (Tasks):**
    1. The user initiates a mission with a high-level goal.
    2. NCC creates a "Master Narrative" object.
    3. As subagents spawn, NCC injects "Chapter" metadata into their context.
    4. Subagents report progress; NCC updates the "Master Narrative."
    5. A Claude Dispatch worker picks up a task and inherits the verified narrative context from the NCC.

## 4. Design & Architecture
* **System Flow:**
    `[Gemini Subagent] <-> [NCC] <-> [Claude Dispatch Worker]`
* **APIs / Interfaces:**
    * `NCC.syncChapter()`: Synchronizes a narrative chapter across framework boundaries.
    * `NCC.getNarrativeContext()`: Provides the summarized narrative for a new subagent.
* **Data Storage/State:**
    * Narrative Graph stored in the Blackboard with parent-child lineage.

## 5. Alternatives Considered
* **Framework-Specific Synchronization:** Rejected as it lacks cross-framework interoperability.
* **Manual Context Injection:** Rejected due to human error and high latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Narrative fragments must be signed to prevent "Story Injection" by rogue subagents.
* **Observability:** Narrative drift is monitored in the "Subagent Lineage Explorer."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
