# Design Doc: Recursive Lease Validator (RLV)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In heterogeneous agent meshes, "Trust Fragmentation" occurs when sub-delegations cross framework boundaries (e.g., from an OpenClaw specialist to a Claude Code teammate). Without a standardized way to pass hardware-attested authority, subagents often lose necessary capabilities, or worse, inherit excessive ones.

RLV solves this by mandating a "Recursive Chain of Authority." Every sub-lease or task delegation must carry a cryptographically signed lineage that RLV validates against the root user intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent "Capability Dropping" during cross-framework handoffs.
    * Enforce monotonic privilege reduction (sub-leases can never exceed parent authority).
    * Provide hardware-attested (TPM/SEP) proof of intent lineage.
* **Non-Goals:**
    * Generic identity management (RLV focus is on *capability leases*).
    * Automated intent generation (RLV validates, it does not create).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer
* **Primary Goal:** Ensure that a "Write" capability granted to a lead agent is only sub-leased to a specific specialist for a limited duration and scope.
* **The Happy Path (Tasks):**
    1. Parent Agent requests a "Mission Lease" for filesystem access.
    2. Parent Agent sub-delegates a "File Edit" task to a Specialist Agent.
    3. Specialist Agent presents the "Sub-Lease" token to RLV.
    4. RLV recursively verifies the signature chain back to the user's Mission Root.
    5. RLV authorizes the tool call only if the sub-lease is mathematically nested within the parent lease's scope.

## 4. Design & Architecture
* **System Flow:**
    * [Root Intent] -> [Master Lease] -> [Sub-Delegation] -> [Sub-Lease Token] -> [RLV Verification] -> [Capability Grant].
* **APIs / Interfaces:**
    * `POST /v1/auth/leases/sub`: Generates a restricted sub-lease token.
    * `POST /v1/auth/leases/validate`: Recursively validates a lease chain against the hardware root.
* **Data Storage/State:**
    * State is managed via **Hardware-Locked Mission Leases (HLML)** and the **Mission-Root Continuity Provider (MRCP)**.

## 5. Alternatives Considered
* **Session-Wide Permissions:** Rejected as it violates the Principle of Least Privilege for autonomous sub-missions.
* **Centralized IAM:** Rejected to maintain low-latency, "Local-Only by Default" operation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RLV is the primary defense against "Identity Shadowing" and "Recursive Intent Poisoning."
* **Observability:** Integrated with the **Subagent Lineage Explorer** for real-time visualization of authority chains.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
