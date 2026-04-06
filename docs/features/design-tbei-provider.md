# Design Doc: Task-Bound Ephemeral Identity (TBEI) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI swarms move toward horizontal collaboration (Agent Teams), the risk of "Identity Squatting" has increased. Subagents from disparate frameworks (Claude Code, OpenClaw, AutoGen) often share a common mission root but operate on different tasks. Without task-level isolation, a compromised subagent could attempt to claim or modify sibling tasks on the shared blackboard, leading to intent drift or data exfiltration.

The Task-Bound Ephemeral Identity (TBEI) Provider addresses this by issuing hardware-attested identity tokens that are cryptographically bound not just to a mission, but to a specific Task UUID on the blackboard.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a local identity service that issues hardware-attested, task-specific tokens.
    * Mandate TBEI verification for all blackboard task claims and modifications.
    * Provide sub-millisecond token issuance and validation.
    * Neutralize "Identity Squatting" in horizontal swarms.
* **Non-Goals:**
    * Providing long-term identity persistence (tokens are ephemeral by design).
    * Replacing framework-level identity (TBEI acts as a mission-bound overlay).

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialist Subagent (e.g., OpenClaw DB Specialist)
* **Primary Goal:** Securely claim a "Database Migration" task from the shared blackboard without being able to influence sibling tasks.
* **The Happy Path (Tasks):**
    1. Parent agent (Mission Root) posts a task to the blackboard with a unique UUID.
    2. Subagent requests to "claim" the task.
    3. TBEI Provider issues a token bound to `(MissionRootID, TaskUUID, SubagentIdentity)`.
    4. Subagent provides the TBEI token to the Blackboard Hub.
    5. Blackboard Hub verifies the token and grants the subagent write access to the specific task entry.
    6. Subagent attempts to modify a different task; the Blackboard Hub rejects the request due to task-UUID mismatch in the TBEI token.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent] -->|Request Task Token| B[TBEI Provider]
        C[Blackboard Hub] -->|Task Metadata| B
        D[TPM/Secure Enclave] -->|Hardware Signature| B
        B -->|TBEI Token| A
        A -->|Claim Task + Token| C
        C -->|Verify Token| E[Identity Validator]
        E -->|Success| F[Grant Task Access]
    ```
* **APIs / Interfaces:**
    * `tbei.IssueToken(taskID TaskUUID, identity HardwareIdentity) -> Token`: Issues a task-bound token.
    * `tbei.VerifyToken(token Token, taskID TaskUUID) -> bool`: Validates a token against a specific task.
* **Data Storage/State:**
    * **Active Token Registry:** A sub-millisecond, in-memory cache of issued and active TBEI tokens, indexed by Task UUID.

## 5. Alternatives Considered
* **Mission-Level Scoping:** Rejected as it allows subagents to access any task within the mission, leading to "Identity Squatting."
* **Short-Lived JWTs:** Used as the underlying transport for TBEI tokens, but enhanced with hardware attestation and task-ID binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Token issuance must be hardware-attested to prevent subagents from "self-minting" unauthorized task tokens.
* **Observability:** Integrated with the "Mission Lineage Tracker" for auditing task-to-identity mappings.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Evolving from Mission-Root Lineage Attestation (2026-06-06) to support granular task-bound identity.
