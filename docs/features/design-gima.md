# Design Doc: GC-Immune Memory Anchor (GIMA)

## Background
As AI agent frameworks scale to handle 1M+ token context windows, "Context-Window Garbage Collection" (CWGC) has become a mandatory optimization to prevent memory overflow and performance degradation. However, aggressive CWGC risks evicting "Silent Anchors"—the implicit instructions, guardrails, and behavioral rules injected early in an agent's lifecycle.

## Problem Statement
When mission-root constraints or core behavioral guardrails are evicted from active memory, the agent experiences "GC Fragility." This can lead to uncontrolled tool usage, semantic drift, or policy violations as the agent forgets its primary directives. We need a way to mark critical context fragments as strictly immune to garbage collection.

## Proposed Solution: GC-Immune Memory Anchor (GIMA)
GIMA provides an advanced pinning mechanism within the MCP Any framework that allows specific context fragments to be anchored securely in memory.
1.  **Pinning API**: Agents and the mission-root can use the `GIMA.Anchor(fragment, priority)` API to lock context shards into the attention window.
2.  **Memory Protection Ring**: GIMA creates an isolated, protected buffer within the larger context window that the underlying model's CWGC algorithm is explicitly instructed to bypass or treat as read-only.
3.  **Lifecycle Integration**: GIMA anchors are tied to the agent's task lifecycle. Once a specific sub-task or mission finishes, the associated anchors are gracefully released to prevent permanent memory bloat.

## Alternatives Considered
-   **Periodic Re-injection**: Re-injecting the guardrails every N turns. *Rejected* due to token inefficiency and the risk of polluting the reasoning trace with repetitive instructions.
-   **External Vector DB Retrieval**: Fetching rules dynamically from a local RAG setup. *Rejected* because latency and context-switching overhead are too high for sub-millisecond behavioral guardrail enforcement.

## Security Implications
-   **Memory Exhaustion (DoS)**: A malicious or malfunctioning subagent could attempt to anchor excessive amounts of text, starving the primary context window. **Mitigation**: GIMA will enforce strict quota limits on the total token count of anchored fragments per subagent and require mission-root attestation for high-priority anchors.
-   **Immutable Malicious Payloads**: If a bad actor manages to anchor a malicious instruction, it becomes permanently active for the duration of the session. **Mitigation**: All GIMA anchoring requests must pass through the Semantic Integrity Bridge and be signed by a trusted Hardware-Locked Mission Lease (HLML).
