# Design Doc: Resident Integrity Monitor (RIM)
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
Hardware-Attested Boot (TPM) ensures an agent starts in a clean environment, but it does not protect against "Delayed Payload" attacks that tamper with the sandbox *after* execution has begun. The Resident Integrity Monitor (RIM) provides continuous, hardware-bound verification that the agent's environment remains in a "Persistent Integrity" state.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform periodic, randomized integrity checks of the agent's sandbox (filesystem, memory, environment variables).
    * Bind integrity signals to hardware security modules (TPM/Secure Enclave).
    * Provide a "Sandbox Persistence Proof" (SPP) required for "Trust Leases" (LFTA).
    * Forcefully terminate sessions if integrity drift is detected.
* **Non-Goals:**
    * Preventing the initial compromise (handled by the Pre-Flight Sandbox Validator).
    * General host-level system monitoring.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a long-running "Research Agent" hasn't had its local configuration swapped via a symlink-race attack 10 minutes after boot.
* **The Happy Path (Tasks):**
    1. Agent is booted with a "Deterministic Boot Manifest."
    2. RIM initializes a hardware-locked monitor for the session.
    3. Every 60 seconds (randomized), RIM validates the SHA-256 hashes of the sandbox's critical configuration handles.
    4. RIM issues a time-bound "Sandbox Persistence Proof."
    5. The A2A Messaging Hub requires a valid SPP before authorizing any "Trust Lease" for tool calls.
    6. If a symlink-race is detected, RIM immediately revokes all active leases and kills the agent process.

## 4. Design & Architecture
* **System Flow:**
    `[TPM/SEP] <-> [RIM] -> [Persistence Proof] -> [Trust Lease Manager]`
* **APIs / Interfaces:**
    * `IntegrityMonitor`: `VerifyState() (SPP, error)`, `GetStatus() MonitorStatus`
    * `SPP`: A cryptographically signed token containing the hardware timestamp and state hash.
* **Data Storage/State:**
    * Stores "Known Good" state manifests in hardware-protected memory regions.

## 5. Alternatives Considered
* **Continuous Polling**: Rejected due to high CPU overhead. RIM uses kernel-level file-watchers combined with randomized hardware-attestation sweeps.
* **Software-Only Integrity**: Rejected as it can be bypassed by an attacker who has achieved sandbox escape.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RIM's heartbeats must be non-predictable to prevent attackers from "timing" their payloads.
* **Observability:** Real-time integrity health is displayed in the "Resident Integrity Status Widget."

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation.
* **2026-04-18:** Optimization for "Resident Persistence Proofs" inspired by Claude Code's latest stability updates. Introducing a "Unified Persistence Broker" pattern to allow swarm-wide sharing of hardware-bound integrity signals, reducing the per-agent attestation tax.
* **2026-04-19:** Integration with Distributed Trust Leases (LFTA). RIM now acts as the authoritative "Lease Guard," providing the continuous hardware-attestation signals required to maintain LFTA token validity. Detection of any integrity drift (e.g., config hook modification) now triggers an immediate global revocation of all active trust leases.
