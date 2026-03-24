# Design Doc: Attention-Boundary Governance (ABG)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms scale horizontally and deep speculative branching becomes common, the integrity of the attention window has become a critical vulnerability. Current attention pinning (HAAL) ensures local consistency but fails to prevent 'Attention-Splicing' (CVE-2026-71002)—a technique where malicious subagents inject high-entropy 'nonsense' shards that force the parent to evict mission-root anchors, effectively 'blinding' the supervisor.

Attention-Boundary Governance (ABG) evolves simple pinning into an active, authoritative service that performs real-time structural analysis of the LLM context window to detect and block 'Attention-Splicing' attempts.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement active structural analysis of the LLM context window to detect Attention-Splicing.
    * Mandate 'Mission-Root Attention Locking' to ensure primary intents remain at the zero-tier attention layer.
    * Provide real-time interdiction of high-entropy noise shards designed to evict critical anchors.
    * Integrate with the Predictive State Purging (PSP) adapter to prevent attention-window flooding.
* **Non-Goals:**
    * Directly managing LLM inference (handled by providers).
    * Enforcing low-level transport security (handled by the Named-Pipe/WebSocket layer).
    * Sanitizing binary state (handled by the WASM-BSH Sanitizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Protect the structural integrity of the supervisor's attention window against malicious entropy injection.
* **The Happy Path (Tasks):**
    1. Parent Agent generates mission-root fragments.
    2. ABG service 'locks' these fragments to the zero-tier attention layer via HAAL headers.
    3. Specialist Agent proposes speculative reasoning shards.
    4. ABG service performs structural analysis on the combined context before it reaches the LLM.
    5. If a malicious shard attempts to 'splice' noise to evict a locked anchor, the ABG service blocks the shard and revokes the subagent's capabilities.
    6. The supervisor's attention remains anchored to the mission root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Context Fragment Stream] --> B[ABG Hub]
        B --> C[Attention Structure Validator]
        C --> D[Entropy Entropy Monitor]
        D --> E{Integrity Verified?}
        E -- Yes --> F[Lock Anchors & Forward to LLM]
        E -- No --> G[Interdict Splicing & Alert Root]
        H[Mission-Root Anchors] --> C
        I[PSP Utility Scores] --> D
    ```
* **APIs / Interfaces:**
    * `abg.ValidateContext(fragments []Fragment) -> bool`: Performs structural and entropy analysis.
    * `abg.LockMissionRoot(missionToken) -> bool`: Pins primary intent fragments.
* **Data Storage/State:**
    * **Attention Map Cache:** A persistent, hardware-attested map of currently 'locked' fragments and their associated mission roots.

## 5. Alternatives Considered
    * **Passive Attention Pinning:** Rejected as it can be overcome by extreme entropy flooding (CVE-2026-71002).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ABG service must utilize Hardware-Attested Identity (DTAI/TIT) to verify the origin of every context fragment.
* **Observability:** Integrated with the 'Context Attention Monitor' for real-time visualization of attention-window health.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
