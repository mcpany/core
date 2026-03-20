# Design Doc: Attention-Locked Speculative Buffer (ALSB)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
The disclosure of "Speculative Attention Hijacking" (CVE-2026-71002) reveals a critical vulnerability where speculative agent branches can be coerced into pre-fetching sensitive context from unrelated mission branches. ALSB provides a hardware-bound isolation layer for speculative context, ensuring that predicted attention maps cannot be used for cross-mission data exfiltration.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically isolated buffers for speculative reasoning.
    * Enforce hardware-bound attention locking for speculative fragments.
    * Prevent cross-mission pre-fetching via predicted attention maps.
    * Support zero-latency commit of speculative results upon attestation.
* **Non-Goals:**
    * Managing non-speculative (committed) state.
    * Encrypting the entire model attention space.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Execute speculative tool calls in Branch B without risking exposure of private data in Branch A.
* **The Happy Path (Tasks):**
    1. The agent spawns a speculative reasoning branch.
    2. ALSB initializes an isolated, hardware-bound buffer for the branch.
    3. Speculative pre-fetching occurs within the locked buffer scope.
    4. An attempt is made to access an attention fragment from another mission root.
    5. ALSB detects the unauthorized attention mapping and blocks the pre-fetch.
    6. Upon successful attestation of the branch, ALSB atomically commits the buffer to the primary state.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        SB[Speculative Branch] -->|Pre-fetch Request| ALSB[ALSB]
        ALSB -->|Check Attention Lock| HAAL[Hardware Attention Lock]
        HAAL -->|Authorized| Buffer[Locked Speculative Buffer]
        HAAL -->|Unauthorized| Block[Block & Alert Hijacking]
        Buffer -->|On Attestation| Commit[Commit to Mission Root]
    ```
* **APIs / Interfaces:**
    * `POST /v1/speculative/buffer/init`: Initialize a locked speculative buffer.
    * `POST /v1/speculative/prefetch`: Request an attention-locked pre-fetch.
* **Data Storage/State:**
    * Buffers are resident in secure memory enclaves with per-fragment attention tags.

## 5. Alternatives Considered
* **Full Branch Isolation:** Rejected due to the overhead of creating complete environment clones for every speculative path.
* **Passive Post-hoc Auditing:** Rejected as it cannot prevent the data exfiltration during the pre-fetch window.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attention locks are hardware-attested and mission-root specific.
* **Observability:** Metrics on "Speculative Hijacking Attempts" visualized in the Connectivity & Security Dashboard.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
