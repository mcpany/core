# Design Doc: Autonomous Context Compaction
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
As AI agents move towards complex multi-agent swarms (e.g., Anthropic's new "Agent Tool" patterns and OpenClaw swarms), the sheer volume of tool-call history, progress updates, and metadata quickly exhausts the LLM's context window. This leads to performance degradation, increased costs, and "context-drift" where the agent loses sight of the primary goal. MCP Any, acting as the universal gateway, is perfectly positioned to manage this history autonomously.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically summarize and compact tool-call history based on session length and token pressure.
    * Implement "Lossless Metadata Stripping" (removing redundant JSON-RPC headers/wrappers).
    * Provide "Goal-Preserving Compression" that prioritizes keeping the high-level intent and final tool results.
    * Enable configurable compaction policies (e.g., "Aggressive" vs "Conservative").
* **Non-Goals:**
    * Modifying the core LLM prompt (this is a history-management middleware).
    * Permanent deletion of history (original traces should remain available for auditing).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Developer
* **Primary Goal:** Maintain agent stability during a 100+ turn autonomous task without manually managing context windows.
* **The Happy Path (Tasks):**
    1. Agent swarm initiates a long-running session via MCP Any.
    2. MCP Any monitors the cumulative token count of the tool-call history.
    3. When a threshold is reached (e.g., 80% of window), the Compactor triggers.
    4. The Compactor replaces 20 intermediate "Search" tool outputs with a single "Consolidated Findings" summary.
    5. The Agent continues execution with a lean, highly relevant context window.

## 4. Design & Architecture
* **System Flow:**
    - **Token Monitor**: Intercepts all tool results and tracks session size.
    - **Compaction Engine**: Uses a lightweight local LLM or rule-based summarizer to process historical chunks.
    - **Context Injector**: Modifies the history returned to the client in subsequent `get_context` calls.
* **APIs / Interfaces:**
    - Middleware Hook: `CompactionMiddleware`
    - Config: `compaction_threshold`, `compaction_strategy: "summarize" | "strip_metadata" | "last_n"`
* **Data Storage/State:** original traces stored in SQLite; compacted view served to the agent.

## 5. Alternatives Considered
* **Client-Side Management:** Rejected because it forces every agent framework (OpenClaw, CrewAI, etc.) to reimplement complex compaction logic.
* **Simple Truncation (FIFO):** Rejected because it often deletes the initial "Instruction" or "Goal" which is critical for agent alignment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The compactor must ensure that security-critical events (like MFA approvals) are never compacted away.
* **Observability:** The UI "Context Dashboard" will visualize "Original Size" vs "Compacted Size" and highlight summarized blocks.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
