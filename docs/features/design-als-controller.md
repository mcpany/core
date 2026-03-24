# Design Doc: ALS (Attention Limit System) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of multi-agent swarms, "Context-Window Ghosting" (CVE-2026-71002) has emerged as a critical vulnerability. Malicious or hallucinating subagents can "flood" the context window with high-entropy tokens, forcing the primary orchestrator to lose sight of the "Mission Root" intent. MCP Any needs a mechanism to enforce **Attention Sovereignty** by capping the token-usage and reasoning-depth of subagents.

The ALS Controller acts as a real-time monitor and enforcer for attention quotas within the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor real-time token consumption per subagent session.
    * Implement "Attention Quotas" that restrict subagent context-window footprints.
    * Automatically terminate reasoning loops that exceed the assigned quota.
    * Provide telemetry for attention utilization across the swarm.
* **Non-Goals:**
    * Performing semantic analysis of tokens (handled by the AID Hub).
    * Managing cloud provider billing directly (handled by PBRB).
    * Modifying model internal weights or attention mechanisms.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Code Debugger" subagent from consuming 90% of the context window during a deep reasoning cycle.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns a subagent with a metadata header `X-MCP-Attention-Quota: 2000`.
    2. Subagent begins reasoning and making tool calls.
    3. ALS Controller tracks cumulative token count in the bidirectional stream.
    4. Subagent hits the 2000-token limit.
    5. ALS Controller injects a `MISSION_EXHAUSTED` interrupt signal and terminates the sub-session.
    6. Orchestrator receives the truncated trace and maintains its own 8000-token mission-root context.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] <-> [ALS Middleware] <-> [Model Provider]`
    The ALS Middleware sits in the message pipeline, incrementing a thread-safe counter for every token in the stream (reasoning monologue + tool output).
* **APIs / Interfaces:**
    * `SetAttentionQuota(SessionID, MaxTokens)`
    * `GetAttentionUtilization(SessionID) -> float`
    * New header: `X-MCP-Attention-Limit: <int>`
* **Data Storage/State:**
    * In-memory session-bound counters for active turns.
    * Persisted stats in the Blackboard for long-running missions.

## 5. Alternatives Considered
* **Client-side Enforcement:** Rejected because subagents can bypass their own limits if compromised or hallucinating.
* **Periodic Polling:** Rejected due to latency; real-time streaming interception is required to prevent "Ghosting" bursts.
* **Model-Native Gating:** Rejected because cloud providers do not expose granular sub-context limits to users.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Limits are cryptographically bound to the mission-root token; subagents cannot update their own quotas.
* **Observability:** Prometheus metrics exported via the Reasoning Telemetry Exporter (`mcp_attention_utilization_ratio`).

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
