# Design Doc: Team-Aware Context Pruning (TACP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale horizontally (e.g., Claude Code Agent Teams), redundant reasoning across teammates leads to significant "Cognitive Stall" and excessive token consumption. Teammates often independently re-verify the same logical dependencies, creating a "Coordination Tax" that slows down the entire mission.

MCP Any needs to implement Team-Aware Context Pruning (TACP) via a Shared Attention Registry (SAR) Hub. This allows teammates to share logical conclusions and deduplicate reasoning traces, ensuring that the swarm operates as a cohesive cognitive unit rather than a collection of isolated agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a Shared Attention Registry (SAR) Hub for cross-agent conclusion sharing.
    * Reduce token consumption by at least 30% in high-density teammate swarms.
    * Ensure "GC-Immune" reasoning anchors remain permanent in the attention window across the team.
    * Provide real-time semantic deduplication of reasoning fragments.
* **Non-Goals:**
    * Automatically merging divergent reasoning branches without mission-root approval.
    * Replacing the SRM (Signed Reasoning Monologue) for private subagent thoughts.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Enable a team of 3 agents to collaboratively debug a complex codebase without re-reading the same dependency maps.
* **The Happy Path (Tasks):**
    1. Agent A reasons through the dependency graph of a specific module and publishes the conclusion to the SAR Hub.
    2. Agent B, tasked with a related module, queries the SAR Hub before starting its reasoning loop.
    3. SAR Hub provides the verified conclusion from Agent A, which is injected into Agent B's context as a "GC-Immune" anchor.
    4. Agent B skips the redundant verification step and proceeds directly to its specific task.

## 4. Design & Architecture
* **System Flow:**
    * Agents emit "Reasoning Fragments" to the SAR Middleware.
    * The SAR Hub performs semantic clustering to identify redundant logic.
    * "Team-Level Conclusions" are persisted in the Shared Attention Registry.
    * The Context Shifter injects these conclusions into relevant teammate windows.
* **APIs / Interfaces:**
    * `sar.v1.PublishConclusion(fragment_id, logic_hash, mission_id)`: Registers a reasoning outcome.
    * `sar.v1.QueryAttention(intent_scope, mission_id)`: Retrieves applicable conclusions for a new task.
* **Data Storage/State:**
    * Shared Attention Registry stored in the project-local SQLite Blackboard.
    * Logic hashes indexed for sub-millisecond retrieval.

## 5. Alternatives Considered
* **Global Context Sharing:** Rejected due to context-window overflow. Sharing everything with everyone creates too much noise.
* **Parent-Only Summarization:** Rejected because it creates a bottleneck at the supervisor agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All shared conclusions must be cryptographically signed by the emitting agent and validated against the mission-root manifest.
* **Observability:** The UI will display a "Shared Attention Heatmap" showing deduplication efficiency and token savings.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
