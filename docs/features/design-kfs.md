# Design Doc: Kernel-Layer Fragment Sanitizer (KFS)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The emergence of "Monologue Splicing" vulnerabilities in high-frequency horizontal coordination has exposed a critical gap in current state management. Malicious subagents can "splice" imperative instructions into a peer's internal monologue by exploiting the sub-millisecond window between fragment attestation and memory ingestion.

The Kernel-Layer Fragment Sanitizer (KFS) addresses this by moving Atomic Fragment Sanitization (AFS) from the application layer to the kernel layer (utilizing eBPF or specialized kernel modules). This ensures that every state fragment is semantically validated and scrubbed *at the point of memory copy*, leaving no window for instruction injection.

## 2. Goals & Non-Goals
* **Goals:**
    * Neutralize "Monologue Splicing" and "Ghost Fragment" exploits.
    * Perform semantic validation of binary state fragments at sub-millisecond latency.
    * Implement sanitization at the kernel-to-user memory boundary.
    * Provide hardware-attested proof of fragment integrity.
* **Non-Goals:**
    * Full reasoning-trace analysis (handled by CAH).
    * Modifying the underlying transport protocol (BSH).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Teammate (e.g., Security Auditor Agent)
* **Primary Goal:** Ingest a binary state fragment from a peer agent without the risk of "Spliced" instructions corrupting its internal reasoning.
* **The Happy Path (Tasks):**
    1. The peer agent sends a `MemoryShard` via the BSH Gateway.
    2. The BSH Gateway initiates a `memfd` copy to the target agent.
    3. The KFS (Kernel-Resident) intercepts the memory copy operation.
    4. The KFS performs real-time semantic analysis of the fragment against the mission-root schema.
    5. If a "Splicing" pattern is detected, the kernel faults the copy and notifies the MFH.
    6. If valid, the fragment is atomically scrubbed of unauthorized tokens and delivered to the agent's memory space.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Gateway as BSH Gateway
        participant Kernel as KFS (Kernel Module / eBPF)
        participant Agent as Target Agent Memory
        participant Hub as ARI/CAH Hub

        Gateway->>Kernel: Initiate memfd_copy(fragment_id)
        Kernel->>Kernel: Semantic Fragment Scan (Layer-7)
        Kernel->>Hub: Verify Fragment Signature
        Hub-->>Kernel: Signature Valid
        alt Splicing Detected
            Kernel-->>Gateway: Fault (Injection Prevented)
        else Valid Fragment
            Kernel->>Agent: Atomic Memory Write (Scrubbed)
            Kernel-->>Gateway: Success
        end
    ```
* **APIs / Interfaces:**
    * `RegisterKFSProbe(mission_root, schema_fingerprint)`
    * `GetSanitizationMetrics() -> (fragments_scanned, injections_blocked)`
* **Data Storage/State:**
    * Sanitization rules and schemas are cached in secure kernel-bound memory.

## 5. Alternatives Considered
* **User-Space Sanitization (AFS):** Rejected because the processing window (50ms+) allows for TOCTOU and splicing exploits.
* **WASM-Bound Sanitization:** Useful for complex schemas, but KFS is required for the low-level memory protection and sub-millisecond latency required for high-frequency swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** KFS operates with root-level isolation. Rules are cryptographically locked to the mission root.
* **Observability:** All kernel-level faults and sanitization events are exported to the Distributed Mesh Resilience (DMR) Hub.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
