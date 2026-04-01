# Design Doc: Recursive Lease Validator (RLV)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of Mission-Bound Hardware Leases (MBHL) provided a mechanism for hardware-attested capability limits. However, as agent swarms grow in depth, parent agents often delegate sub-missions with the full scope of their own hardware leases. This creates a "Lease Creep" risk where a specialized subagent (e.g., a file parser) might inherit broad administrative privileges (e.g., full shell access) simply because its parent possessed them.

The Recursive Lease Validator (RLV) is designed to act as a mandatory "Lease Narrower." It ensures that every sub-delegation carries a cryptographically signed hardware lease that is a strictly defined subset of the parent's authority, maintaining the principle of least privilege throughout the agentic tree.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically narrowing TPM-signed hardware leases for all sub-delegated missions.
    * Validating the complete "Chain of Lease" for every tool call across the swarm.
    * Neutralizing "Lease Creep" by blocking subagent requests that exceed parent-narrowed boundaries.
    * Integrating with the existing HLML Provider to ensure hardware-bound revocation.
* **Non-Goals:**
    * defining the high-level security policies (handled by the Policy Firewall).
    * Managing the transport layer for leases (handled by the A2A Messaging Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Ensure a "Refactor Specialist" subagent can only write to a specific directory, even if the "Project Lead" agent has full project-wide write access.
* **The Happy Path (Tasks):**
    1. Parent Agent (Lease A: `fs:write:/project/*`) initiates a sub-delegation to Specialist Agent.
    2. RLV intercepts the delegation and generates a Narrowed Lease (Lease B: `fs:write:/project/src/utils/*`).
    3. RLV cryptographically signs Lease B and binds it to the Specialist's hardware identity.
    4. Specialist Agent attempts to write to `/project/Makefile`.
    5. RLV/HLML Provider detects that `/project/Makefile` is outside the boundaries of Lease B.
    6. The tool call is interdicted, and a "Lease Boundary Violation" alert is raised.

## 4. Design & Architecture
* **System Flow:**
    `Parent Intent` -> `RLV (Lease Narrowing)` -> `Narrowed Hardware Lease` -> `Subagent Execution` -> `RLV/HLML (Enforcement)`.
* **APIs / Interfaces:**
    * `rlv.NarrowLease(parentLease, subIntent) -> NarrowedLease`: Generates a subsetted lease.
    * `rlv.VerifyChain(leaseChain) -> boolean`: Verifies the cryptographic parentage of a lease.
* **Data Storage/State:**
    * **Lease Lineage Cache:** In-memory store of active lease parent-child relationships for fast-path validation.

## 5. Alternatives Considered
* **Explicit Policy Declarations**: Rejected because it adds significant complexity to agent prompts; an automated infrastructure-level validator is more reliable.
* **Process-Level Sandboxing (gVisor)**: RLV is complementary; gVisor handles OS isolation, while RLV handles agentic capability isolation at the hardware-attestation level.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Chain of Lease" must be immutable and verifiable by any node in the mesh.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time visualization of the lease tree and narrowing events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
