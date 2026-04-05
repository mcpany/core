# Design Doc: Kernel-Resident TOCTOU Protection (KRTP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Mesh-Resident Memory Mapping" (MRMM) in high-performance agent frameworks like OpenClaw, agent nodes are increasingly sharing state via direct memory-mapped regions across high-speed interconnects. While this drastically reduces coordination latency, it introduces a critical Time-of-Check Time-of-Use (TOCTOU) vulnerability. A malicious or compromised specialist agent could mutate a state fragment in the shared buffer *after* it has been validated by the security gate but *before* it is ingested by the recipient agent's reasoning engine.

MCP Any needs to solve this by providing a kernel-mediated mechanism that "freezes" state fragments at the moment of validation, ensuring the integrity of the data handed off to the model.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide hardware-enforced immutability for shared memory regions immediately following security validation.
    *   Support sub-millisecond transition from "Writeable" (by producer) to "Read-Only" (for consumer) and "Verified" (by MCP Any).
    *   Integrate with existing WASM-BSH sanitizers to trigger "Freeze" signals.
*   **Non-Goals:**
    *   Implementing a new distributed filesystem or memory manager (we leverage existing MRMM/memfd).
    *   Providing data encryption at rest (handled by other Strategic Pillars).

## 3. Critical User Journey (CUJ)
*   **User Persona:** High-Frequency Teammate Mesh Orchestrator
*   **Primary Goal:** Share a 500MB database schema fragment between an OpenClaw "DB Architect" and a Claude Code "Backend Dev" without risk of mid-flight corruption.
*   **The Happy Path (Tasks):**
    1.  Producer agent writes state to a `memfd`-backed shared buffer.
    2.  MCP Any intercepts the "Handoff" signal and invokes the WASM-BSH Sanitizer.
    3.  Upon successful validation, MCP Any issues a `pkey_mprotect` syscall to assign a unique Memory Protection Key (MPK) to the region.
    4.  The hardware kernel freezes the region as Read-Only for all agents except the recipient (who gets Read-Access).
    5.  Recipient agent ingests the "Frozen" fragment with a hardware-attested integrity receipt.

## 4. Design & Architecture
*   **System Flow:**
    [Producer Agent] --(Write)--> [Shared Mem]
                                      |
    [MCP Any Validator] <---(Signal)---'
          |
    (WASM Scan Pass)
          |
    [KRTP Controller] --(syscall: pkey_mprotect)--> [Hardware MMU]
                                                       |
    [Consumer Agent] <---(Read-Only Map)---------------'

*   **APIs / Interfaces:**
    *   `POST /v1/mesh/memory/freeze`: Validates and locks a specific memfd segment.
    *   `GET /v1/mesh/memory/attestation`: Returns a hardware-signed proof that the segment was frozen at timestamp T.
*   **Data Storage/State:**
    *   State is stored in `memfd_create` anonymous file descriptors.
    *   Lock state is managed via Linux Kernel MPK (Memory Protection Keys).

## 5. Alternatives Considered
*   **Buffer Copying:** Rejecting due to the 150ms+ latency tax for large fragments (Token Storms).
*   **User-Space Mutexes:** Rejected because they can be bypassed by malicious subagents with direct memory access.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** We assume the hardware kernel and MMU are the root of trust. MPKs prevent even root-level subagents from modifying the buffer once the key is restricted.
*   **Observability:** All "Freeze" events are logged with SHA-256 hashes and timestamped in the Blackboard Lineage.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
