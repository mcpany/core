# Design Doc: Active Context Refresh (ACR) Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents engage in long-running reasoning sessions, the increasing density of the context window often leads to "Instruction Eviction," where the original system guardrails and mission-root anchors are pushed out of the model's active attention window. This "GC Fragility" results in semantic drift, where agents lose their behavioral constraints and mission focus.

The ACR Middleware addresses this by acting as an authoritative "Attention Guardian." It monitors the context window in real-time and automatically re-injects mission-critical guardrails at the attention-head of the LLM context, ensuring that core behavioral guardrails remain permanent without requiring model-side prompting.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time monitoring of LLM context window occupancy.
    * Automatically re-inject "Mission-Root" anchors and safety guardrails.
    * Implement "Semantic Token Compression" to minimize the overhead of refresh cycles.
    * Ensure mission-root sovereignty across long reasoning chains.
* **Non-Goals:**
    * Modifying the underlying LLM's attention mechanism.
    * Providing a general-purpose RAG solution for all history fragments.

## 3. Critical User Journey (CUJ)
* **User Persona:** Long-Running Autonomous Agent Supervisor
* **Primary Goal:** Maintain agent behavioral consistency during a 1M+ token reasoning session.
* **The Happy Path (Tasks):**
    1. The supervisor defines a "Mission-Root" manifest with core guardrails.
    2. The ACR Middleware calculates the "Attention Expiry" for the current session.
    3. As reasoning progresses, ACR detects that core anchors are nearing the eviction boundary.
    4. ACR triggers a "Refresh Cycle," re-injecting the anchors at the top of the context.
    5. The agent continues reasoning with refreshed behavioral constraints.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B[ACR Middleware]
        B --> C{Context Threshold Met?}
        C -- Yes --> D[Re-inject Mission-Root Anchors]
        C -- No --> E[Proceed to LLM]
        D --> F[Semantic Token Compression]
        F --> E
        E --> G[LLM Response]
    ```
* **APIs / Interfaces:**
    * `RefreshContext(session_id string, anchors []Anchor)`
    * `SetAttentionThreshold(threshold float64)`
* **Data Storage/State:**
    * Uses the `Mission-Root Registry` to store immutable anchor fragments.

## 5. Alternatives Considered
* **Model-side Pinning**: Rejected due to inconsistent support across different LLM providers (Gemini vs. Claude).
* **Manual Re-prompting**: Rejected due to high latency and the risk of the agent misinterpreting the re-prompt as a new instruction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ACR only re-injects hardware-attested anchors verified by the Mission-Root Registry.
* **Observability:** Logs "Refresh Events" and the associated token savings from Semantic Token Compression.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
