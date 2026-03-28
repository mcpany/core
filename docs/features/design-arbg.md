# Design Doc: Adaptive Role-Bound Gating (ARBG)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
With the release of Claude Code Opus 4.7, AI agents have gained the ability to autonomously swap roles (e.g., from Researcher to Implementer) within a horizontal team. While this reduces coordination latency, it introduces a significant security risk: "Capability Squatting." A specialized agent could transition into a role with higher privileges (like `admin` or `shell_executor`) without explicit parent-agent or user verification, effectively bypassing current Zero-Trust boundaries.

ARBG solves this by acting as the authoritative "Role Broker" for the Universal Agent Bus. It mandates that every role transition be backed by a hardware-attested "Role Card" that defines the exact capabilities and lifecycle of that role.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all inter-agent role-swap signals in horizontal swarms.
    * Validate role transitions against hardware-attested "Role Cards."
    * Forcefully revoke capabilities associated with the previous role upon transition.
    * Provide a cryptographic audit trail of all role-swapping events.
* **Non-Goals:**
    * Defining the LLM's internal persona or behavior.
    * Managing non-agent human user roles (out of scope for agent infrastructure).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Allow a research teammate to transition into an implementer role and gain write access to a specific git branch, but ONLY after proving mission-root alignment.
* **The Happy Path (Tasks):**
    1. Researcher Agent completes data gathering and issues a `request_role_swap(new_role="Implementer")` signal.
    2. MCP Any ARBG interceptor pauses the request and retrieves the Mission-Root's authorized Role Manifest.
    3. ARBG requests a hardware-attested identity proof from the agent.
    4. ARBG validates that "Researcher" -> "Implementer" is a valid transition for this specific mission branch.
    5. ARBG issues a new, time-bound Capability Token for the Implementer role and revokes the Researcher's vector-search tokens.
    6. The agent resumes as Implementer with limited write access.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>ARBG: request_role_swap(Implementer)
        ARBG->>FSI_Provider: Verify Hardware Identity
        FSI_Provider-->>ARBG: Identity Validated
        ARBG->>Policy_Engine: Check Transition Permitted (Researcher -> Implementer)
        Policy_Engine-->>ARBG: Permitted
        ARBG->>Token_Mint: Issue Implementer Capability Token
        ARBG->>Token_Mint: Revoke Researcher Tokens
        ARBG-->>Agent: Role Swapped (Implementer)
    ```
* **APIs / Interfaces:**
    * `POST /v1/roles/swap`: Initiates a role transition request.
    * `GET /v1/roles/manifest`: Retrieves the allowed role transition graph for the active mission.
* **Data Storage/State:**
    * Role transitions and active Role Cards are stored in the Shared KV Store (Blackboard) under the `system:roles` namespace, protected by hardware-bound locks.

## 5. Alternatives Considered
* **Static Role Binding:** Assigning a fixed role at spawn time. Rejected because it fails to support the efficiency gains of modern horizontal swarms (Claude Code 4.7 patterns).
* **Pure LLM-Gated Swapping:** Letting the agent announce its role change. Rejected as it provides zero security guarantees against rogue subagents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Role Cards are TPM-signed and linked to the Mission-Root. Unauthorized transition attempts trigger immediate agent suspension via the MSSQ (Machine-Speed Swarm Quarantine).
* **Observability:** Every transition is logged in the `Action-Chain Sovereignty Monitor` (ACSM) with a reference to the triggering reasoning fragment.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
