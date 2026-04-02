# Design Doc: Root Attention Segment (RAS) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLM context windows exceed 1M+ tokens, aggressive "Context-Window Garbage Collection" (CWGC) is becoming standard to maintain performance. However, this has introduced "Instruction Eviction" vulnerabilities, where core behavioral guardrails (the mission root) are pruned from the model's attention window during high-entropy reasoning flushes.

The Root Attention Segment (RAS) Provider implements the Gemini AACP standard, providing a hardware-locked, privileged memory region that persists instructions through force-flushes, ensuring that agents never "forget" their security constraints.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-locked context pinning for mission-root instructions.
    * Ensure that "Silent Anchors" remain permanent in the attention window regardless of CWGC aggressive pruning.
    * Support AACP-compliant privileged attention segments across heterogeneous LLM providers.
* **Non-Goals:**
    * Managing general-purpose RAG memory (handled by ContextEngine).
    * Increasing the physical size of the context window; RAS optimizes the *density* and *persistence* of existing tokens.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Agent Developer
* **Primary Goal:** Ensure behavioral guardrails (e.g., "Never exfiltrate .env files") survive even when the agent is processing a 50,000-line codebase.
* **The Happy Path (Tasks):**
    1. Developer defines mission-root guardrails in the Mission Manifest.
    2. RAS Provider flags these instructions for "Privileged Pinning."
    3. During LLM inference, the RAS Provider injects these anchors into the hardware-locked Root Attention Segment.
    4. As the agent reasons and the context window fills, the LLM provider triggers a CWGC flush.
    5. Standard context fragments are pruned, but the RAS Provider prevents the eviction of the RAS segment.
    6. The agent continues reasoning with the guardrails still in its primary attention window.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Guardrails] --> B[RAS Flagger]
        B --> C[RAS Provider]
        C --> D{LLM Provider API}
        D -->|AACP Header| E[Privileged Attention Segment]
        E --> F[Persistent Reasoning Anchors]
    ```
* **APIs / Interfaces:**
    * `ras.PinToRoot(fragmentID, priority) -> Status`: Locks a context fragment into the RAS.
    * `ras.GetPersistenceProof(fragmentID) -> HardwareProof`: Attests that the segment was not evicted during the last flush.
* **Data Storage/State:**
    * **RAS Manifest:** TPM-signed list of active fragments in the privileged attention tier.

## 5. Alternatives Considered
* **Constant Re-Injection of Instructions:** Rejected due to "Token Inflation." Repeating guardrails in every prompt significantly increases costs and can confuse the model's logic. RAS provides a single, permanent anchor.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RAS segments are immutable once pinned for the duration of the mission phase.
* **Observability:** Visualized via the "GC-Immune Anchor Visualizer" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
