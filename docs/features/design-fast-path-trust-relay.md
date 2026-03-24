# Design Doc: Fast-Path Trust Relay (FPTR)
**Status:** Draft
**Created:** 2026-06-30

## 1. Context and Scope
Multi-hop delegation (e.g., Lead Agent -> Specialist -> Sub-Specialist) is the backbone of modern swarm intelligence. However, the requirement for full hardware-attested handshakes at every hop has introduced a "Handshake Fatigue" latency penalty (up to 400ms in deep chains). This delay discourages granular delegation and forces agents to "over-reason" in a single session, leading to context bloat and hallucination.

The **Fast-Path Trust Relay (FPTR)** is a performance-optimizing security service for MCP Any. It facilitates the propagation of hardware-attested trust leases across infinite delegation hops, maintaining the full cryptographic strength of the "Mission Root" without the per-hop signature overhead.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce multi-hop delegation latency by 75%+.
    *   Propagate hardware-attested trust strength across 10+ delegation hops.
    *   Implement "Trust Leases" that automatically expire upon task completion.
    *   Neutralize "Handshake Fatigue" while maintaining MLE (Mission-Locked Execution) compliance.
*   **Non-Goals:**
    *   Bypassing the initial hardware handshake (required for mission-root initiation).
    *   Storing private keys on subagents (trust is relayed via signed tokens).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Delegate a complex data transformation through 3 specialist agents without incurring a 1-second "Security Tax."
*   **The Happy Path (Tasks):**
    1.  The Lead Agent completes a full hardware handshake with MCP Any, initiating the Mission Root.
    2.  MCP Any issues a "Root Trust Lease" (RTL) via the FPTR.
    3.  Lead Agent delegates to Specialist A, passing the RTL.
    4.  Specialist A presents the RTL to the FPTR for a "Sub-Lease."
    5.  FPTR verifies the RTL's hardware lineage and issues a sub-millisecond sub-lease.
    6.  Specialist A completes its task and delegates to Specialist B using the sub-lease.
    7.  The delegation chain continues with sub-10ms security overhead at each step.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        TPM[Hardware TPM] -->|Initial Sign| Root[Mission Root Agent]
        Root -->|Handshake| FPTR[Fast-Path Trust Relay]
        FPTR -->|Issue Lease| L1[Level 1 Lease]
        L1 --> AgentA[Specialist A]
        AgentA -->|Delegate + L1| AgentB[Specialist B]
        AgentB -->|Present L1| FPTR
        FPTR -->|Verify Lineage| L2[Level 2 Lease]
        AgentB -->|Execute with L2| Tools[MCP Tools]
    ```
*   **APIs / Interfaces:**
    *   `fptr.InitiateMission(hardwareProof) -> RootLease`: Starts the trust chain.
    *   `fptr.RelayTrust(parentLease, targetIdentity) -> SubLease`: Extends the chain.
    *   `fptr.VerifyLease(lease) -> bool`: Validates a lease for tool access.
*   **Data Storage/State:**
    *   "Trust Lineage Tree" stored in a protected segment of the Blackboard, mapping lease IDs to hardware-attested parents.

## 5. Alternatives Considered
*   **Persistent WebSocket Connections**: Rejected because agents often run in ephemeral processes (Claude Code) or across different containers, making persistent sockets unreliable.
*   **Shared Token Registry**: Rejected as it lacks the cryptographic parent-child binding required to prevent "Token Hijacking."

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All leases are time-bound and cryptographically linked to the hardware Inode of the mission manifest.
*   **Observability:** Integrated with the "Mesh-Resident Lineage Tracker" to visualize the trust delegation path.

## 7. Evolutionary Changelog
*   **2026-06-30:** Initial Document Creation.
