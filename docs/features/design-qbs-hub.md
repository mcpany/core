# Design Doc: Quorum-Bound Summarization (QBS) Hub
**Status:** Draft
**Created:** 2026-07-06

## 1. Context and Scope
As AI agent swarms move toward long-running sessions with context windows exceeding 1M tokens, "context compaction" (summarization) becomes a standard operational requirement to maintain performance. However, existing "Priority-Aware" compaction strategies often exhibit "Summarization Ghosting" (Mission-Root Erasure), where over-optimized summarizer agents inadvertently drop core mission constraints or security boundaries to save tokens.

MCP Any needs to act as the authoritative "Compaction Arbiter" by introducing a consensus-based validation layer for all state summarization events. This ensures that the mission-root intent remains sovereign and untampered during the lifecycle of the swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-agent quorum for context summarization.
    * Explicitly protect "Mission-Root" fragments from being dropped or altered during compaction.
    * Provide hardware-attested approval tokens for the final compressed state.
* **Non-Goals:**
    * Replacing existing summarization engines (e.g., OpenClaw ContextEngine). QBS Hub acts as a validator, not a generator.
    * Real-time text-to-speech or non-textual trace summarization in this phase.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Ensure that a 48-hour autonomous coding mission doesn't lose its "No External API calls" security constraint due to aggressive context summarization.
* **The Happy Path (Tasks):**
    1. The primary Agent Framework (e.g., Claude Code) triggers a context compaction event.
    2. MCP Any intercepts the proposed summary via the ContextEngine Adapter.
    3. The QBS Hub identifies "Mission-Root" fragments within the proposed summary.
    4. The QBS Hub spawns or delegates to an "Independent Auditor" agent to compare the summary against the original mission manifest.
    5. The Mission-Root agent and Auditor agent both sign off on the summary.
    6. MCP Any commits the validated summary to the Blackboard and hardware-attests the new state.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Framework] -> [Compaction Request] -> [QBS Hub] -> [Auditor Agent Validation] -> [Quorum Approval] -> [Blackboard Commit]`
* **APIs / Interfaces:**
    * `mcp.compact_context(original_fragments, proposed_summary) -> QuorumToken`
    * `mcp.attest_summary(summary_hash, agent_signature)`
* **Data Storage/State:**
    * Summarization events are logged in the Blackboard Versioning Hub.
    * Quorum signatures are stored as hardware-attested metadata bound to the session.

## 5. Alternatives Considered
* **Heuristic-Based Weighting**: Assigning high weights to mission-root tokens. Rejected because weights can still be over-indexed by summarizers during high-entropy reasoning bursts.
* **Human-in-the-Loop (HITL) for every summary**: Rejected due to prohibitive latency and "Approval Fatigue" in autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Auditor agents must have zero knowledge of session-bound secrets, validating only the semantic consistency of constraints.
* **Observability:** Real-time "Summarization Drift" scores will be visible in the Context Shifting Timeline.

## 7. Evolutionary Changelog
* **2026-07-06:** Initial Document Creation.
