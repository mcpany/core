# Design Doc: Memory-Fence Sanitization (MFS) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The discovery of "Shadow-Attestation" vulnerabilities revealed that nanosecond timing drift between the system clock and the Trusted Platform Module (TPM) can be exploited to inject "Ghost Fragments" into hardware-signed traces.

The Memory-Fence Sanitization (MFS) Provider secures the "Moment of Attestation." It enforces hardware-level memory fences and mandatory monotonic clock synchronization before any reasoning fragment is cryptographically finalized. This ensures that the state being signed is physically and temporally consistent with the hardware root of trust.

## 2. Goals & Non-Goals
* **Goals:**
    * Force monotonic clock synchronization (`PTP/NTP` alignment) before every TPM signature operation.
    * Implement `mfence/sfence` (or equivalent kernel primitives) to ensure context isolation during signing.
    * Neutralize "Shadow-Attestation" exploits by closing the 50ns-100ns timing-drift window.
* **Non-Goals:**
    * General-purpose memory encryption (focus is on the attestation pipeline).
    * Eliminating all timing side-channels (focus is specifically on Shadow-Attestation injection).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor
* **Primary Goal:** Verify that a hardware-signed reasoning trace from a remote specialist agent hasn't been tampered with via timing-drift injection.
* **The Happy Path (Tasks):**
    1. A specialist agent generates a reasoning fragment and requests a TPM signature.
    2. The MFS Provider intercepts the request and issues a hardware memory fence.
    3. MFS performs a monotonic clock-drift check against the hardware secure timer.
    4. MFS "flushes" the reasoning buffer and synchronizes the system clock.
    5. The fragment is signed by the TPM.
    6. The auditor verifies the trace, which now includes an "MFS-Synchronized" attestation bit, proving temporal integrity.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Request[Signature Request] --> MFS[MFS Middleware]
        MFS --> Fence[Hardware Memory Fence]
        Fence --> Sync[Monotonic Clock Sync]
        Sync --> Flush[Buffer Flush]
        Flush --> TPM[TPM Sign]
        TPM --> Result[Signed Fragment]
    ```
* **APIs / Interfaces:**
    * `x-mcp-mfs-status`: Header indicating the MFS synchronization state of a fragment.
* **Data Storage/State:**
    * Monotonic counter values are stored in kernel-protected memory.

## 5. Alternatives Considered
* **Higher Precision TPMs:** Rejected as it requires expensive hardware upgrades for all nodes.
* **Pure Software Jitter:** Rejected as it doesn't solve the underlying timing-drift exploit, only masks it.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MFS itself must be hardware-attested. A node cannot claim "MFS-Protected" status without proving Inode-Pinning of the MFS driver.
* **Observability:** Clock-drift variations are logged for anomaly detection.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
