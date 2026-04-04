# Design Doc: HAMR-Compliant Recovery Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agents, particularly those in high-stakes environments like "Agent Teams" (Claude Code), can occasionally diverge from their intended mission parameters due to reasoning drift or context injection. While "Deterministic Sandbox Recovery" (DSR) provides basic rollback capabilities, it lacks hardware-bound integrity.

The HAMR-Compliant Recovery Hub introduces **Hardware-Attested Mission Rollbacks (HAMR)**. It allows the mission-root to forcefully reset the execution environment to a TPM-signed snapshot when reasoning-path watermarking (RPW) or mission-manifest verification signals a violation. This ensures that the agent cannot "persist" through a compromise.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized hub for managing hardware-signed project snapshots.
    * Trigger automated rollbacks based on "Reasoning-Divergence" signals from ALT or SRM.
    * Forcefully revoke all mission-bound capability leases (HLML) upon rollback initiation.
    * Provide a verifiable audit trail of mission resumption events and hardware signatures.
* **Non-Goals:**
    * General-purpose system backup (focus is strictly on the agent's project-local environment).
    * Real-time reasoning steering (HAMR is a discrete recovery mechanism).

## 3. Critical User Journey (CUJ)
* **User Persona:** Mission-Root Supervisor
* **Primary Goal:** Reset a subagent's environment after it attempts an unauthorized action detected by the Attention-Locked Tooling (ALT).
* **The Happy Path (Tasks):**
    1. A subagent reasoning trace shows it is being driven by "Injected Context" (detected by ALT).
    2. The ALT middleware interdicts the tool call and sends a "Divergence Alert" to the HAMR Hub.
    3. The HAMR Hub verifies the current session state against the hardware-attested manifest.
    4. The Hub triggers a forceful environment rollback to the last known-good TPM-signed snapshot in the Shadow-FS.
    5. The "Orphaned Lease Interdictor" automatically revokes all active HLML leases.
    6. The mission resumes with a clean state, notifying the supervisor of the intervention.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Subagent] -->|Reasoning Trace| ALT[ALT Middleware]
        ALT -->|Divergence Alert| HAMR[HAMR Hub]
        HAMR -->|Verify| Manifest[Mission Manifest]
        HAMR -->|Revoke| HLML[Lease Interdictor]
        HAMR -->|Rollback| SFS[Shadow-FS Snapshot]
        SFS -->|Resume| Agent
    ```
* **APIs / Interfaces:**
    * `POST /v1/recovery/snapshot`: Create a TPM-signed environment snapshot.
    * `POST /v1/recovery/rollback`: Initiate a hardware-attested mission rollback.
    * `GET /v1/recovery/history`: Audit log of all recovery interventions.
* **Data Storage/State:**
    * TPM-signed block-level snapshots stored in secure, mission-bound buffers.
    * Monotonic "Rollback Counters" to prevent snapshot replay attacks.

## 5. Alternatives Considered
* **Software-only DSR:** Rejected because it can be bypassed by subagents exploiting kernel-level persistence or TOCTOU vulnerabilities.
* **Continuous Checkpointing:** Rejected due to excessive I/O overhead; HAMR prioritizes "Milestone-based" hardware signing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Rollbacks and snapshots must be cryptographically anchored to the user's hardware identity.
* **Observability:** Visualized via the "Mission Resumption Manager" and "Durable Session Explorer" UI components.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
