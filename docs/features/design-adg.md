# Design Doc: Attention-Density Guard (ADG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the adoption of 1M+ token context windows, a new Denial-of-Service (DoS) pattern has emerged: Context-Window Flooding (CWF). In multi-agent swarms, a malicious or compromised subagent can inject high-entropy "noise" (redundant reasoning, large blobs of irrelevant text) to force the eviction of mission-critical instructions or "Mission Root" anchors from the LLM's active attention window. MCP Any needs to actively protect the "Attention Sovereignty" of the primary mission.

## 2. Goals & Non-Goals
* **Goals:**
    * Protect mission-critical context fragments from being evicted from the attention window.
    * Detect and throttle subagents exhibiting high "Reasoning Entropy" with low mission utility.
    * Implement hardware-bound "Attention Anchors" that LLMs are forced to prioritize.
* **Non-Goals:**
    * Arbitrarily limiting the total context window size (scalability is still desired).
    * Modifying the underlying LLM's attention mechanism (we operate at the gateway/orchestration layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Maintain mission-root sovereignty even if a specialist subagent attempts to flood the context with 500k tokens of noise.
* **The Happy Path (Tasks):**
    1. The user defines a "Mission Root" with specific "Immutable Anchors" in MCP Any.
    2. A swarm of 10 subagents begins processing a large dataset.
    3. Subagent 7 (compromised) attempts a CWF attack by outputting 200k tokens of high-entropy garbage.
    4. The ADG middleware detects the entropy spike and compares it against the Mission Root's utility score.
    5. ADG "pins" the Immutable Anchors using hardware-bound headers (e.g., `x-mcpany-attention-priority: high`).
    6. ADG dynamically prunes or summarizes the noise from Subagent 7 before it reaches the parent agent's context.
    7. The parent agent continues to reason correctly, anchored by the protected instructions.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Output] --> B{ADG Entropy Scanner}
        B -->|High Entropy / Low Utility| C[Noise Pruning / Summarization]
        B -->|Normal Flow| D[Context Assembler]
        E[Mission Root Anchors] --> F[Attention Pinning Provider]
        F --> D
        D --> G[Parent LLM Input]
    ```
* **APIs / Interfaces:**
    * `x-mcpany-attention-priority`: A new header used in tool-calling and A2A messages to signal priority.
    * `ADG.Prune(fragment)`: Internal method for semantic noise reduction.
* **Data Storage/State:**
    * ADG maintains a real-time "Entropy Scorecard" for every active subagent in the session.

## 5. Alternatives Considered
* **Static Token Quotas:** Rejected because some valid tasks (e.g., large file analysis) require high token counts; static quotas are too blunt.
* **LLM-Based Summarization:** Rejected as the primary defense because the summarization itself could be "poisoned" by the noise it's trying to reduce.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Attention priority is hardware-attested; subagents cannot "elevate" their own noise priority.
* **Observability:**
    * "Attention Eviction" events are logged as high-severity security alerts.
    * Users can visualize the "Attention Map" in the MCP Any UI to see what's driving the agent.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
