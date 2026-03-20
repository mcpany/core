# Design Doc: Mesh-Resident Governance Oracle (MRGO) Adapter

**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope

As agent swarms scale horizontally, the "Mission-Root Bottleneck" has become a critical performance and reliability risk. Currently, every high-stakes policy decision must be escalated back to the lead agent, adding latency and increasing the risk of mission-root exhaustion. OpenClaw v3.3 introduces the MRGO to allow decentralized, local teammate coordination for policy arbitration. MCP Any needs to host this oracle to enable autonomous, secure governance within the horizontal mesh.

## 2. Goals & Non-Goals

* **Goals:**
    * Provide hardware-attested "Governance Quorums" for horizontal agent teams.
    * Enable local teammate policy arbitration without mission-root escalation.
    * Ensure all mesh-resident decisions remain cryptographically bound to the mission-root intent.
    * Mitigate the risk of "Governance Drifting" via periodic alignment heartbeats.
* **Non-Goals:**
    * Replacing the mission-root as the ultimate authority for the entire swarm.
    * Managing tool-specific permissions (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator (e.g., Claude Code Team Lead)
* **Primary Goal:** Reach a "Governance Quorum" among 3 specialized teammates to authorize an ambiguous tool call without waking the lead agent.
* **The Happy Path (Tasks):**
    1. A specialist teammate initiates a "Policy Arbitration" request via the MRGO.
    2. The MRGO broadcasts the request to authenticated peers in the mesh.
    3. Peers perform independent reasoning trace analysis and submit TPM-signed approval/denial tokens.
    4. The MRGO collects the tokens and verifies they reach the required quorum threshold.
    5. The tool call is executed using the mesh-resident "Governance Token."
    6. The decision is asynchronously reported to the mission-root for lineage auditing.

## 4. Design & Architecture

* **System Flow:**
    ```mermaid
    graph TD
        S[Specialist Teammate] -->|Policy Request| MRGO[MRGO Adapter]
        MRGO -->|Broadcast| Peers[Teammate Mesh]
        Peers -->|TPM-Signed Token| MRGO
        MRGO -->|Verify Quorum| Quorum[Governance Quorum]
        Quorum -->|Valid| Execute[Execute Tool Call]
        Quorum -->|Invalid| Deny[Deny Tool Call]
        Execute -->|Audit Trail| MR[Mission Root]
    ```
* **APIs / Interfaces:**
    * `POST /v1/governance/arbitrate`: Initiate a local policy arbitration request.
    * `POST /v1/governance/quorum/vote`: Submit a hardware-attested vote for a pending request.
    * `GET /v1/governance/quorum/status`: Monitor the status of an active governance quorum.
* **Data Storage/State:**
    * Quorum states and votes are stored in a transient, hardware-locked buffer within the MRGO.

## 5. Alternatives Considered

* **Centralized Lead Governance:** Rejected due to high latency and mission-root exhaustion risks in large swarms.
* **Shared-State (Blackboard) Voting:** Rejected as the Blackboard is vulnerable to "Mesh-Split" (CVE-2026-82001) divergence.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** All votes are hardware-bound. "Governance Drifting" is mitigated by mandatory AIA alignment heartbeats.
* **Observability:** Full audit trail of all mesh-resident decisions exported to the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog

* **2026-06-18:** Initial Document Creation.
