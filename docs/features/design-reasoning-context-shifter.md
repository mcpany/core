# Design Doc: Reasoning-Bound Context Shifter
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
As agents move toward "Deep Reasoning" (e.g., OpenClaw v2.5), the amount of active context required to maintain mission focus exceeds the efficient token window of most LLMs. "Reasoning-Bound Context Shifting" allows agents to dynamically mount and unmount context fragments based on their current reasoning branch. However, without strict governance, this leads to **"Branch Contamination"** (leakage of discarded hypothetical state) and **"Context Amnesia"** (loss of critical mission parameters). MCP Any must act as the authoritative synchronizer for these shifts.

## 2. Goals & Non-Goals
* **Goals:**
    * Synchronize context shifting logic across multiple agent frameworks.
    * Enforce "Branch Purity" by isolating the Blackboard and memory during hypothetical reasoning.
    * Provide "Parental Re-Attestation" before a sub-branch's state is merged into the primary intent.
* **Non-Goals:**
    * Implementing the LLM's internal attention mechanism.
    * Managing the agent's long-term vector memory (handled by external RAG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep-Reasoning Swarm Orchestrator
* **Primary Goal:** Explore 3 divergent debugging paths (A, B, C) for a production outage without Path A's false assumptions contaminating Path C.
* **The Happy Path (Tasks):**
    1. Parent agent initiates 3 parallel sub-intents via UACO.
    2. MCP Any creates isolated "Shadow Blackboards" for each branch.
    3. As agents "shift" context relevant to their path, MCP Any validates that the "Primary Mission Fragment" remains pinned.
    4. Path A is discarded; MCP Any purges its Shadow Blackboard.
    5. Path C succeeds; MCP Any requests Parental Re-Attestation before merging Path C's state into the primary mission context.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Shift Request] -> [Context Shifter Middleware] -> [Branch Purity Validator] -> [Shadow Blackboard/Context Shard]`
* **APIs / Interfaces:**
    * `context/shift`: Mounts a specific shard while preserving the "Mission-Critical" base.
    * `context/checkpoint`: Creates a snapshot of the current reasoning state for rollback.
* **Data Storage/State:**
    Uses memory-mapped "Shadow Blackboards" for ephemeral branch state.

## 5. Alternatives Considered
* **Monolithic Context:** Simple but leads to token storms and context-smearing.
* **Purely Client-Side Shifting:** Rejected because it allows framework-specific bugs to cause cross-agent state contamination in shared swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Branch Purity" is essential to prevent a compromised subagent from injecting malicious "hypothetical" state into the parent.
* **Observability:** Provides a "Branch History" view in the UI to trace how reasoning paths diverged.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
