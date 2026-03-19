# Design Doc: Attention-Locked Sovereignty (ALS) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the discovery of "Context-Window Ghosting" (CVE-2026-71002), it is evident that high-entropy semantic noise can be used to evict critical safety instructions and mission-root anchors from an LLM's attention window. This vulnerability allows attackers to bypass an agent's security boundaries by simply flooding the context with irrelevant but non-malicious data.

MCP Any needs to solve this by providing a mechanism to "pin" or "lock" safety-critical intents at the attention layer, ensuring they remain resident regardless of the input volume from subagents or tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound "Attention-Locking" headers for LLM requests.
    * Provide a mechanism for the Mission Root to pre-declare "Pinned Instructions."
    * Automatically detect and mitigate "Context-Window Ghosting" attempts via real-time attention-utilization analysis.
* **Non-Goals:**
    * Modifying the internal architecture of third-party LLMs.
    * Managing non-safety-critical context fragments (this remains the role of the ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Ensure that "Deny" rules and data-handling constraints remain effective even when a specialist agent retrieves 10MB of documentation.
* **The Happy Path (Tasks):**
    1. The Mission Root pre-attests a set of "Immutable Constraints" (e.g., "Never exfiltrate PII").
    2. MCP Any wraps these constraints in ALS-locking headers.
    3. A subagent performs a massive document retrieval tool call.
    4. MCP Any's ALS Controller monitors the context expansion and ensures the ALS headers are re-injected or prioritised in the attention window.
    5. The LLM continues to respect the "Deny" rule despite the retrieval bloat.

## 4. Design & Architecture
* **System Flow:**
    [Mission Root] -> [ALS Controller: Wrap with Attention-Locking Headers] -> [LLM Gateway]
    [Subagent Tool Output] -> [ALS Controller: Attention-Utilization Analysis] -> [LLM Gateway: Prioritize ALS Headers]
* **APIs / Interfaces:**
    * `SetAttentionLock(MissionID, FragmentID, Priority)`
    * `OnContextExpansion(SessionID, DeltaSize)`
* **Data Storage/State:**
    * Hardware-bound registry of pinned fragment hashes.
    * Real-time attention-utilization scoreboard per session.

## 5. Alternatives Considered
* **Frequent Re-prompting:** Rejected due to excessive token cost and reasoning latency.
* **Aggressive Summarization:** Rejected as it may inadvertently lose subtle safety nuances required for zero-trust enforcement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALS headers must be cryptographically bound to the Mission Root to prevent subagents from "unpinning" their own constraints.
* **Observability:** Log "Eviction Risk" events when high-entropy noise approaches the attention limit.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
