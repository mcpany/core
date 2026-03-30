# Design Doc: Speculative Intent Sanitizer (SIS)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
With the introduction of Speculative Intent Buffering (SIB) in OpenClaw v3.5, agents can now begin reasoning on predicted intents before official attestation. While this reduces MTTC, it introduces a critical vulnerability: "Speculative Intent Poisoning." Malicious subagents can inject instructions into these speculative buffers that bypass current post-hoc validation.

SIS provides a real-time semantic sanitization layer specifically for speculative fragments, ensuring that speculative reasoning remains bound to the verified mission root and cannot be hijacked by "poisoned" predictions.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, pre-commitment semantic analysis of all speculative reasoning fragments.
    * Neutralize "Speculative Intent Poisoning" attempts by subagents.
    * Enforce strict intent-alignment for any fragment entering the speculative buffer.
    * Integrate with the OpenClaw SIB transport.
* **Non-Goals:**
    * Replacing the final mission-root attestation.
    * Validating non-speculative, fully attested fragments (handled by AID).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Teammate Swarm
* **Primary Goal:** Use speculative buffering to reduce latency without risking mission-root hijacking by a specialist subagent.
* **The Happy Path (Tasks):**
    1. Parent agent enables Speculative Intent Buffering for a high-priority sub-task.
    2. Subagent A begins providing speculative reasoning fragments to the mission-root buffer.
    3. SIS intercepts each fragment in real-time.
    4. SIS performs a semantic drift check against the parent's verified mission-root intent.
    5. No drift is detected; the fragment is allowed to influence the parent's speculative reasoning loop.
    6. (Failure Path): Subagent B attempts to "smuggle" an unauthorized intent into the speculative buffer. SIS detects the drift and immediately purges the buffer, revoking Subagent B's speculative privileges.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Fragment] --> B[SIS Interceptor]
        B --> C[Semantic Alignment Engine]
        C --> D{Aligned with Mission Root?}
        D -- Yes --> E[Speculative Buffer]
        D -- No --> F[Buffer Purge & Revoke]
        E --> G[Speculative Reasoning Engine]
    ```
* **APIs / Interfaces:**
    * `SanitizeSpeculativeFragment(ctx, fragment, missionRoot) (isSafe, sanitizedFragment, error)`
    * `SetSpeculativeConfidenceThreshold(ctx, threshold) error`
* **Data Storage/State:** SIS maintains a lightweight, ephemeral state of the current speculative branch, which is flushed upon mission commitment or failure.

## 5. Alternatives Considered
* **Disabling Speculative Buffering:** Rejected due to the 40% performance penalty (MTTC increase).
* **Delayed Sanitization:** Rejected because the parent's reasoning is influenced *during* ingestion, making post-hoc sanitization ineffective.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SIS ensures that "Speculative Trust" is never permanent and is continuously verified.
* **Observability:** SIS rejection events and semantic drift scores are logged for swarm forensics.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
