# Design Doc: ALS (Attention Limit System) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agentic swarms become more complex, a significant bottleneck has emerged: "Attention Sharding" and "Context-Window Ghosting." Rogue subagents or deep reasoning loops can capture the primary agent's attention by flooding the context window with high-entropy noise, effectively "ghosting" the original mission intent.

The ALS Controller is designed to provide active, hardware-attested attention management. It moves MCP Any from a passive transport layer to a reasoning-aware gateway that enforces attention quotas, ensuring that the mission-root remains pinned and subagents cannot monopolize the cognitive window.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce real-time "Attention Quotas" (token limits) per subagent session.
    * Provide hardware-attested "Attention Locking" for mission-critical intent fragments.
    * Detect and terminate "Ghosting Loops" before they capture the orchestrator's attention.
    * Export attention-utilization metrics for observability and budget enforcement.
* **Non-Goals:**
    * Modification of the underlying LLM's attention mechanism or weights.
    * Semantic summarization of context (handled by ContextEngine).
    * General-purpose rate limiting (handled by separate ratelimit middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (e.g., Senior AI Architect)
* **Primary Goal:** Prevent a specialized "Code Auditor" subagent from capturing the primary "Lead Developer" agent's attention during a massive refactoring task.
* **The Happy Path (Tasks):**
    1. Orchestrator configures an Attention Policy for the swarm with a 4k token quota per subagent.
    2. The Lead Developer agent delegates a refactoring task to the Auditor.
    3. The Auditor begins generating a high-entropy reasoning trace.
    4. ALS Controller monitors the Auditor's token footprint in real-time.
    5. When the Auditor hits 90% of its quota, ALS injects a "Low Attention" signal.
    6. If the Auditor exceeds the quota without reaching convergence, ALS terminates the session and purges its ghost state from the Blackboard.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] -> [ALS Middleware] -> [Attention Monitor] -> [Context Window]
    ALS intercepts every interaction, calculating the "Attention Impact" based on token count and semantic entropy.
* **APIs / Interfaces:**
    * `SetAttentionQuota(agentID string, quota int)`
    * `LockAttentionFragment(fragmentID string, priority int)`
    * `GetAttentionMetrics(sessionID string) AttentionReport`
* **Data Storage/State:**
    Attention quotas and current utilization are stored in a high-speed, in-memory Shard-Aware Mailbox.

## 5. Alternatives Considered
* **Passive Monitoring:** Rejected because it allows context poisoning to happen before intervention.
* **Fixed Context Windows:** Rejected because modern swarms require dynamic, intent-aware attention allocation.
* **LLM-Native Gating:** Rejected due to lack of cross-framework standardization and hardware-attestation support.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention quotas are cryptographically bound to hardware-attested mission tokens to prevent subagents from "stealing" attention from peers.
* **Observability:** Metrics are exported via the `Reasoning Telemetry Exporter` to Prometheus for real-time dashboarding.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
