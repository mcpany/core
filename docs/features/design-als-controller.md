# Design Doc: ALS Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of multi-agent swarms, "Context-Window Ghosting" (CVE-2026-71002) has emerged as a critical vulnerability. Malicious or hallucinating subagents can "flood" the context window with high-entropy tokens, forcing the primary orchestrator to lose sight of the "Mission Root" intent. MCP Any needs a mechanism to enforce **Attention Sovereignty** by capping the token-usage of subagents.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor real-time token consumption per subagent session.
    * Implement "Attention Quotas" that restrict subagent context-window footprints.
    * Automatically terminate reasoning loops that exceed the assigned quota.
* **Non-Goals:**
    * Performing semantic analysis of tokens (handled by the AID Hub).
    * Managing cloud provider billing directly (handled by PBRB).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from consuming 90% of the context window.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns a subagent with a metadata header `X-MCP-Attention-Quota: 2000`.
    2. Subagent begins reasoning and tool calls.
    3. ALS Controller tracks cumulative token count in the bidirectional stream.
    4. Subagent hits the 2000-token limit.
    5. ALS Controller injects a `MISSION_EXHAUSTED` interrupt signal and terminates the sub-session.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] <-> [ALS Middleware] <-> [Model Provider]`
    The ALS Middleware increments a thread-safe counter for every token in the bidirectional stream.
* **APIs / Interfaces:**
    New header: `X-MCP-Attention-Limit: <int>`
* **Data Storage/State:**
    In-memory session-bound counters for active turns.

## 5. Alternatives Considered
* **Client-side Enforcement:** Rejected because subagents can bypass their own limits if compromised.
* **Periodic Polling:** Rejected due to latency; real-time streaming interception is required.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Limits are cryptographically bound to the mission-root token.
* **Observability:** Prometheus metrics exported via the Reasoning Telemetry Exporter.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
