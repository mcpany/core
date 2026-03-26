# Design Doc: Mesh Policy Synchronizer (MPS)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As agent swarms scale horizontally across multiple frameworks (Claude Code teams,
OpenClaw specialists), maintaining a consistent security posture becomes
increasingly difficult. Parallel agent teams often exhibit "Policy Drift,"
where divergent security guardrails lead to unauthorized tool calls or
inconsistent handling of sensitive data.

The Mesh Policy Synchronizer (MPS) provides an authoritative hub for real-time
synchronization of security guardrails across the agent mesh. It ensures that
all teammates operate under a unified, hardware-attested policy set.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-10ms synchronization of security policies across the mesh.
    * Support framework-agnostic policy formats (Rego/CEL).
    * Implement hardware-attested policy signing to prevent subagent tampering.
    * Enable real-time "Policy Heartbeats" to detect and resolve drift.
* **Non-Goals:**
    * Implementing the policy enforcement engine itself (handled by Policy Firewall).
    * Managing global policies outside the local mesh scope.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Administrator for Multi-Agent Swarms.
* **Primary Goal:** Update a "Data Redaction" policy and have it propagate to
  10 parallel subagents in under 100ms.
* **The Happy Path (Tasks):**
    1. The Admin updates the policy via the MPS API.
    2. MPS validates the policy and signs it with the Mesh-Root TPM.
    3. MPS broadcasts the update to all active teammate nodes via the T2T bus.
    4. Teammate agents receive the heartbeat and update their local cache.
    5. The Policy Firewall on each node begins enforcing the new rules instantly.
    6. MPS receives "Sync Confirmation" from all nodes.

## 4. Design & Architecture
* **System Flow:**
    [Admin/Root] --> [MPS Hub]
    [MPS Hub] --(Signed Policy Update)--> [T2T Encryption Bridge]
    [T2T Bridge] --(Broadcast)--> [Teammate 1..N (Local Cache)]
    [Local Cache] --> [Policy Firewall]
* **APIs / Interfaces:**
    * `POST /v1/mesh/policy/update`: Submit a new mesh-wide policy.
    * `GET /v1/mesh/policy/status`: Check sync status across all nodes.
    * `GET /v1/mesh/policy/heartbeat`: Internal endpoint for node synchronization.
* **Data Storage/State:**
    * Signed policy manifest stored in the Mesh-Resident Attestation Registry.

## 5. Alternatives Considered
* **Centralized Policy Retrieval:** Rejected due to the 50-100ms latency of
  per-call remote lookups.
* **Static Config Files:** Rejected because they cannot be updated in real-time
  during an active mission.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Policies are hardware-bound and session-specific.
  Unauthorized attempts to modify policies trigger a mesh-wide lockdown.
* **Observability:** Sync latency and drift events are tracked in the Mesh
  Policy Editor UI.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
