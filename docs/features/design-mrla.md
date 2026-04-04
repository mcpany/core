# Design Doc: Mesh-Resident Lease Arbiter (MRLA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of Mission-Bound Hardware Leases (MBHL) in Claude Code and similar frameworks has introduced a new failure mode in horizontal Agent Teams: **Lease Deadlocks**. Specialized subagents are often blocked from interdependent tasks because their hardware-attested capability leases are too rigid and lack overlapping scope.

The Mesh-Resident Lease Arbiter (MRLA) is an authoritative governance service in MCP Any that facilitates dynamic, hardware-attested lease expansion between parallel teammates.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism for agents to request temporary, mission-bound capability expansions.
    * Utilize hardware-attested (TPM/Secure Enclave) tokens to verify the legitimacy of expansion requests.
    * Automate the resolution of capability conflicts in multi-agent meshes without requiring human intervention.
* **Non-Goals:**
    * Replacing existing per-mission manifests (it dynamicallly extends them).
    * Bypassing User-Defined Security Policies (expansions must remain within the "Policy Ceiling").

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Agent Mesh Orchestrator
* **Primary Goal:** Resolve a capability deadlock where Teammate A (Filesystem Specialist) and Teammate B (Shell Specialist) need cross-capabilities to complete a shared refactor task.
* **The Happy Path (Tasks):**
    1. Teammate A encounters a task requiring shell access, which is currently leased only to Teammate B.
    2. Teammate A issues a "Lease Expansion Request" to the MRLA via the Sovereign Node Tunnel.
    3. The MRLA verifies Teammate A's mission-root lineage and hardware attestation.
    4. The MRLA checks the user's Policy Ceiling to ensure shell access is permitted for this mission.
    5. The MRLA issues a temporary, hardware-signed "Capability Graft" to Teammate A.
    6. Teammate A completes the task; the graft automatically expires.

## 4. Design & Architecture
* **System Flow:**
    * **Arbiter Core**: The decision engine that evaluates expansion requests against policies and lineage.
    * **Trust Token Exchange**: Handles the secure handoff of hardware-attested tokens between mesh nodes.
    * **Lease Registry**: Tracks the active state and expiration of all primary and grafted leases.
* **APIs / Interfaces:**
    * `POST /mesh/lease/request`: Initiate an expansion request.
    * `GET /mesh/lease/status`: Query active leases for a mission branch.
* **Data Storage/State:**
    * Persistent storage of mission manifests and policy ceilings in the Mesh-Aware Blackboard.

## 5. Alternatives Considered
* **Centralized Manual Approval**: Rejected due to high latency and "Approval Fatigue" in large swarms.
* **Global Least Privilege**: Rejected as it increases the attack surface if any single agent is compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All expansion requests must be cryptographically linked to a hardware-attested mission-root.
* **Observability:** Detailed audit logs of all lease grafts and expansion denials.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
