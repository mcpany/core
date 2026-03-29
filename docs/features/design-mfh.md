# Design Doc: Mission Forking Hub (MFH)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As agent swarms become more complex and decentralized, the traditional "delegation" model is proving insufficient for maintaining strict security boundaries. Current delegation often allows subagents to inherit parent permissions without immutable constraints, leading to "Intent Ghosting" and unauthorized boundary expansion.

The Mission Forking Hub (MFH) introduces a new paradigm: **Mission Forking**. Unlike delegation, a "fork" creates a sub-mission with a dedicated, immutable security policy inherited from the parent. This policy remains physically bounded by the parent's resource and security manifest even if the sub-mission migrates between different mesh nodes.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide an authoritative broker for spawning sub-missions.
    * Ensure inherited security policies are immutable and cryptographically bound to the parent mission.
    * Maintain "Sovereignty-by-Design" across mesh migrations.
    * Support physical resource bounding for forked missions.
* **Non-Goals:**
    * Replacing the existing delegation model for low-trust/low-risk tasks.
    * Managing the internal reasoning of the forked mission (focus is on the boundary).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Lead Architect
* **Primary Goal:** Spawn a specialized data-analysis sub-swarm that can access specific database shards but is physically barred from making network calls or accessing root env vars.
* **The Happy Path (Tasks):**
    1. The Lead Agent identifies a need for specialized sub-tasks.
    2. The Lead Agent requests a "Mission Fork" from the MFH, providing a signed sub-manifest.
    3. The MFH validates the sub-manifest against the Lead Agent's root policy.
    4. The MFH spawns the forked sub-mission with a hardware-attested, immutable security anchor.
    5. The sub-mission executes on a separate mesh node, carrying its immutable policy.
    6. Any attempt by the sub-mission to exceed the forked boundaries is interdicted by local enforcers using the MFH-issued anchor.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Parent as Mission Root (Parent)
        participant MFH as Mission Forking Hub
        participant Mesh as Mesh Node B (Worker)
        participant Enforcer as Local Policy Enforcer

        Parent->>MFH: Request Fork(Sub-Manifest, Hardware-Signature)
        MFH->>MFH: Validate against Parent Root Policy
        MFH-->>Parent: Fork Approved (Fork-Token, Immutable-Policy)
        MFH->>Mesh: Deploy Forked Session (Fork-Token, Policy)
        Mesh->>Enforcer: Load Immutable Policy
        Mesh-->>MFH: Fork Active
        Note over Mesh, Enforcer: Sub-mission bounded by Fork-Token
    ```
* **APIs / Interfaces:**
    * `ForkMission(parent_token, sub_manifest) -> (fork_token, immutable_policy_anchor)`
    * `VerifyForkSovereignty(fork_token) -> (parent_lineage, current_constraints)`
* **Data Storage/State:**
    * Fork lineages and manifests are stored in the hardware-attested Mesh Registry.

## 5. Alternatives Considered
* **Legacy Delegation:** Rejected because it lacks the "Immutable Anchor" required for physical bounding across mesh migrations.
* **Global Security Service:** Rejected to avoid a central bottleneck; MFH focuses on issuing "anchors" that are enforced locally on mesh nodes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Fork tokens are hardware-bound and session-locked. Policies are cryptographically signed by the MFH.
* **Observability:** Every "Fork" event and boundary violation is logged to the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
