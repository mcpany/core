# Design Doc: Hierarchical Mission Lease Orchestrator (HMLO)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of deep agent swarms and distributed meshes, the simple "Point-to-Point" privilege model has reached its limit. "Lease Chaining" (as seen in Claude Code v3.2.1) and "Recursive Tunneling" (OpenClaw v3.7.0) demand a centralized, hierarchical system for managing mission-bound capabilities. Without this, swarms suffer from "Lease Fragmentation" and "Entropy Drift," where sub-agents either lack necessary tools or retain unauthorized privileges.

The Hierarchical Mission Lease Orchestrator (HMLO) is the authoritative service for managing the lifecycle of subsetted, hardware-locked leases across infinite delegation hops.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate "Lease Chaining" where sub-leases are issued from parent MBHLs.
    * Enforce strict "Capability Subsetting" (child privileges must be a subset of parent).
    * Implement "Entropy-Aware Lease Issuance" to prevent delegation during cognitive instability.
    * Provide a unified registry for tracking hierarchical lease lineage across the mesh.
* **Non-Goals:**
    * Defining the specific reasoning logic for tool selection (handled by the Agent).
    * Replacing the underlying TPM attestation; it coordinates the results of that attestation.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Safely delegate a "Write-Access to /src/api" sub-task to a remote edge node while the parent retains "Full FS" access on the primary node.
* **The Happy Path (Tasks):**
    1. Primary Agent initiates a sub-delegation to a remote node.
    2. HMLO verifies the parent's MBHL and current reasoning entropy.
    3. HMLO generates a "Chained Lease" with strictly subsetted permissions (/src/api) and a shorter TTL.
    4. The lease is cryptographically bound to the remote node's hardware identity.
    5. Remote specialist agent executes the task using the subsetted lease.
    6. Upon task completion, HMLO automatically revokes the sub-lease and reconciles the state back to the parent mission-root.

## 4. Design & Architecture
* **System Flow:**
    `Parent MBHL` -> `HMLO Interlock (Entropy Check)` -> `Sub-Lease Generation` -> `Recursive Handshake` -> `Sub-Task Execution` -> `Cascading Revocation`
* **APIs / Interfaces:**
    * `hmlo.IssueSubLease(parentLeaseID, subsetScope, targetNodeID) -> SubLeaseID`: Issues a chained lease.
    * `hmlo.VerifyLineage(leaseID) -> LineageToken`: Verifies the entire path back to the hardware-root.
    * `hmlo.SyncEntropy(agentID, score)`: Real-time update of reasoning entropy for the interlock.
* **Data Storage/State:**
    * **Hierarchical Lease Registry:** Graph database (embedded) tracking lease parentage and active status.
    * **Entropy Buffer:** High-speed cache for real-time agent confidence scores.

## 5. Alternatives Considered
* **Flat Token Lists:** Rejected because they don't support automated cascading revocation or subsetting enforcement.
* **Purely Local Leases:** Rejected because they fail in multi-node mesh environments where lineage must be verified across hops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All sub-leases are signed by the HMLO's hardware key and are physically bound to the target node.
* **Observability:** Visualized in the "Recursive Mission-Root Lineage Visualizer" and "Mission Lease Manager" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
