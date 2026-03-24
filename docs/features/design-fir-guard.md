# Design Doc: Fluid Intent Re-composition (FIR) Guard
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
With the release of OpenClaw 3.4.0, agents can now perform "Fluid Intent Re-composition" (FIR), allowing them to dynamically re-negotiate their mission bounds mid-session. While this increases flexibility, it creates a massive "Bound-Drift" vulnerability where a subagent can gradually escalate its own privileges or expand its scope beyond the user's original intent without a full cold-boot attestation.

MCP Any needs to act as the authoritative "Bound-Arbiter" that anchors these fluid changes to a hardware-attested root manifest, ensuring that any intent expansion is cryptographically validated and user-authorized.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate mid-reasoning intent expansion requests against the hardware-attested mission-root.
    * Enforce a "Delta-Attestation" flow for fluid re-negotiations.
    * Provide a cryptographic proof of the complete intent-mutation lineage.
* **Non-Goals:**
    * Replacing the primary orchestration logic of OpenClaw.
    * Managing the internal reasoning states of subagents (handled by ARI Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent subagents from autonomously expanding their file-system access bounds during a fluid mission update.
* **The Happy Path (Tasks):**
    1. A specialist agent identifies a need to access a new directory (e.g., `/etc/config`) not in its original manifest.
    2. The agent issues a "Fluid Intent Update" request to the FIR Guard.
    3. The FIR Guard intercepts the request and checks it against the TPM-signed "Mission-Root Policy."
    4. If the expansion is within "Pre-Approved Elastic Bounds," the Guard signs the new intent fragment.
    5. If outside bounds, it triggers a hardware-locked HITL (Human-In-The-Loop) re-attestation.
    6. The agent receives the signed "Elastic Bound Token" and continues reasoning.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `FIR Update Request` -> `FIR Guard (Validation against TPM Root)` -> `[Optional] HITL Challenge` -> `Signed Elastic Token` -> `Agent`
* **APIs / Interfaces:**
    * `POST /v1/intent/fluid-expand`: Endpoint for agents to request bound updates.
    * `GET /v1/intent/lineage`: Retrieve the cryptographically signed history of intent mutations.
* **Data Storage/State:**
    * Intent Mutation Log: Append-only, hash-chained log stored in the hardware-protected Blackboard shard.

## 5. Alternatives Considered
* **Cold-Boot Only Enforcement:** Rejected because it destroys the performance gains of OpenClaw's fluid architecture by forcing 2s+ restarts for every minor scope change.
* **Purely Behavioral Monitoring:** Rejected because it is reactive; "Bound-Drift" must be blocked pre-execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All FIR requests must be signed with the agent's session-bound identity token.
* **Observability:** Every intent expansion is logged to the "Sovereignty Audit Dashboard" with a visual diff of the bound changes.

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
