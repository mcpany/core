# Design Doc: Hardware-Attested Resumption Verifier (HARV)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
With the introduction of Atomic Mission Resumption (AMR), agent state can now survive process crashes and cold-boots via hardware-locked snapshots. However, today's market sync revealed "Mission-Root Ghosting," where valid resumption tokens are re-played to hijack a mission frontier.

HARV addresses this by acting as the authoritative state-controller for hardware-attested resumption. It ensures that tokens are not only cryptographically valid but are also hardware-revocable and bound to the absolute current state of the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a hardware-revocable registry for Mission Resumption Tokens.
    * Mandate TPM-bound monotonic counters for every "Atomic Snapshot."
    * Enforce "Single-Use Consistency" where a resumption token is invalidated the moment state is modified.
    * Integrate with the Reasoning Sovereignty Enforcer to verify token provenance.
* **Non-Goals:**
    * Encrypting the entire BSH state (handled by ESB).
    * Providing long-term archival of reasoning history.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect for AI Swarms
* **Primary Goal:** Prevent an attacker from using a previous turn's resumption token to fork or hijack the mission root.
* **The Happy Path (Tasks):**
    1. Agent performs a tool call and requests an Atomic Snapshot.
    2. HARV generates a Resumption Token bound to a TPM Monotonic Counter.
    3. The node restarts.
    4. Agent submits the token for resumption.
    5. HARV verifies the counter value and the hardware signature.
    6. Token is marked as "In-Use."
    7. Agent resumes execution. Any attempt to use the same token again results in a "Monotonic Divergence" error.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>AMR Gateway: Snapshot Request
        AMR Gateway->>HARV: Issue Monotonic Token
        HARV->>TPM: Increment & Sign Counter
        TPM-->>HARV: HARI Signature
        HARV-->>Agent: Monotonic Resumption Token
        Note over Agent, HARV: Restart Event
        Agent->>HARV: Resume (Token)
        HARV->>TPM: Verify Counter Identity
        alt Valid
            HARV-->>Agent: Restore State
        else Stale/Ghosted
            HARV-->>Agent: Sovereignty Failure (Revoked)
        end
    ```
* **APIs / Interfaces:**
    * `HARV_Issue_Token(mission_id, snapshot_hash) -> token`
    * `HARV_Verify_And_Revoke(token) -> boolean`
* **Data Storage/State:**
    * Hardware-bound "Active Token Manifest" stored in the Mission-Root Enclave.

## 5. Alternatives Considered
* **Time-based Expiration**: Rejected because it doesn't prevent sub-millisecond replay or ghosting if the time window is still open.
* **In-Memory Revocation List**: Rejected because it doesn't survive the very restart events AMR is designed to handle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Relies on TPM Monotonic Counters to provide a hardware-guaranteed "Happens-Before" relationship for snapshots.
* **Observability:** Track `ResumptionFailureRate` specifically for `STALE_TOKEN` errors to detect potential ghosting attempts.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
