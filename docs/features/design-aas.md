# Design Doc: Atomic Attention Sharding (AAS)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
The emergence of "Reasoning Entropy Exhaustion" (REE) and "Context-Window Flooding" (CWF) has exposed a fundamental weakness in current agentic context management. Malicious subagents can overwhelm a supervisor agent by injecting high-entropy noise into shared shards, forcing the eviction of the primary mission-root intent from the active context window.

The Atomic Attention Sharding (AAS) Middleware is needed to allow agents to selectively "blind" themselves to non-relevant context fragments at the hardware level, ensuring that mission-critical intent remains protected from noise-driven eviction.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism for agents to selectively mask context fragments based on their relevance to the current sub-task.
    * Prevent "Context-Window Flooding" by automatically pruning high-entropy, low-utility noise.
    * Enforce hardware-locked attention boundaries for mission-critical intent fragments.
    * Support real-time attention-utilization analysis to detect REE attacks.
* **Non-Goals:**
    * Permanently deleting context fragments (masking is session-bound).
    * Replacing the need for summarization (AAS is for attention governance, not compression).

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialized Subagent (e.g., Database specialist)
* **Primary Goal:** Focus on a complex SQL optimization task without being distracted or overwhelmed by high-frequency log noise from a sibling "Logging" agent.
* **The Happy Path (Tasks):**
    1. The Subagent initializes its reasoning branch with an "Attention Manifest" pre-approved by the supervisor.
    2. The AAS Middleware masks all context fragments that do not match the manifest (e.g., excluding sibling log shards).
    3. The agent performs its task with a "Clean" context window containing only the SQL schema and mission-root intent.
    4. If the sibling agent attempts to flood the shared shard with noise, the AAS Middleware blocks the ingestion of these fragments into the specialist's attention layer.
    5. The mission-root intent remains pinned and untainted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Context Ingestion] --> B[AAS Middleware]
        B --> C[Attention Manifest Filter]
        C -->|Relevant| D[LLM Attention Layer]
        C -->|Irrelevant| E[Hardware Masking]
        F[Mission-Root Supervisor] -->|Issue Manifest| B
        G[REE Detector] -->|Trigger Pruning| B
    ```
* **APIs / Interfaces:**
    * `aas.SetAttentionManifest(manifest) -> bool`: Defines the active attention mask for the session.
    * `aas.MaskFragment(fragmentID) -> void`: Manually masks a specific context fragment.
    * `aas.GetAttentionScores() -> map[FragmentID]float`: Returns real-time attention utilization scores.
* **Data Storage/State:**
    * **Attention Manifests:** Hardware-encrypted masks stored in the Mission-Root Enclave.
    * **Fragment Metadata:** Real-time entropy and utility scores for active context fragments.

## 5. Alternatives Considered
* **Soft Attention Weighting:** Rejected as it can be bypassed by model-level prompt injection that forces attention shifts. Hardware-level "blinding" is more deterministic.
* **Aggressive Summarization:** Rejected because it loses the "Atomic" detail needed for complex specialist reasoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AAS manifests must be hardware-signed by the Mission-Root agent. Subagents cannot modify their own attention masks to "see" unauthorized state.
* **Observability:** Integrated with the "Context Attention Monitor" for real-time visualization of masking events and REE noise levels.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
