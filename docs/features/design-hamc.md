# Design Doc: Hardware-Attested Monotonic Clocks (HAMC)
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
Lock-free coordination in horizontal meshes (LFMA) relies on timestamps to resolve conflicts in Conflict-Free Replicated Data Types (CRDTs). However, the disclosure of CVE-2026-41221 reveals that a compromised subagent can manipulate its local system clock offset to "win" conflict resolutions, effectively injecting stale or malicious state into the Shared Task List.

HAMC provides a high-fidelity timing service that utilizes hardware-bound (TPM) monotonic counters to generate cryptographically signed, un-spoofable timestamps. This ensures that temporal coordination remains resilient to clock-drift injection attacks.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested monotonic timestamps for inter-agent coordination.
    * Neutralize "Clock-Drift Injection" attacks in CRDT conflict resolution.
    * Ensure temporal consistency across heterogeneous agent frameworks.
    * Implement sub-millisecond clock attestation for high-frequency swarms.
* **Non-Goals:**
    * Syncing with global NTP time (HAMC focuses on monotonic consistency, not absolute wall-clock time).
    * Replacing local system clocks for non-coordination tasks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Administrator
* **Primary Goal:** Prevent a malicious specialist agent from overwriting a teammate's task completion status by spoofing a "future" timestamp.
* **The Happy Path (Tasks):**
    1. Agent A completes a task and requests a timestamp from the HAMC provider.
    2. HAMC queries the hardware TPM monotonic counter and signs the result.
    3. Agent A submits the task completion with the `HAMC-Token` to the LFMA mesh.
    4. Malicious Agent B attempts to overwrite this with a spoofed "Last-Writer-Wins" update using a back-dated system clock.
    5. LFMA identifies the missing or invalid `HAMC-Token` on Agent B's request and rejects the update.
    6. The Shared Task List remains consistent and secure.

## 4. Design & Architecture
* **System Flow:**
    [Agent Request] ---> [HAMC Provider] ---> [TPM Monotonic Counter]
                               |
                   [Cryptographic Signing Engine]
                               |
                  [Hardware-Attested Timestamp] ---> [LFMA Middleware]

* **APIs / Interfaces:**
    * `GET /v1/time/attestation`: Retrieve a signed hardware timestamp.
    * `X-HAMC-Monotonic-Counter`: Header containing the signed counter value.
* **Data Storage/State:**
    * Monotonic counter state is maintained by the hardware enclave.
    * HAMC provider caches the last signed counter to detect replay attempts.

## 5. Alternatives Considered
* **Centralized NTP Server:** Rejected because NTP traffic is susceptible to MITM drift injection and network latency.
* **Vector Clocks:** Rejected due to the "Identity Shadowing" risk where a compromised agent can fork its own vector history.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Timestamps are cryptographically bound to the hardware identity of the host node.
* **Observability:** Clock-skew and drift anomalies are visualized in the **Local Security Audit Dashboard**.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
