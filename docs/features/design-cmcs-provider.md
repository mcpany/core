# Design Doc: Cross-Mesh Command Sovereignty (CMCS)
**Status:** Draft
**Created:** 2026-05-29

## 1. Context and Scope
As agent swarms move from hierarchical (parent-child) to horizontal (teammate-to-teammate) coordination, the risk of "Teammate Impersonation" where a compromised agent sends malicious mailbox messages to a sibling becomes critical. CMCS provides a hardware-attested "Mesh Token" that binds every inter-teammate command to the mission root and its authorized role.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographically signed "Mesh Tokens" for teammate-to-teammate (T2T) mailboxes.
    * Mandate role-bound capability validation for all mailbox requests.
    * Provide non-repudiable audit logs for all cross-mesh commands.
* **Non-Goals:**
    * Enforcing tool-level permissions (handled by the Policy Firewall).
    * Providing end-to-end encryption for the mailbox content (handled by the T2T Encryption Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Ensure a "Researcher" teammate cannot coerce a "Coder" teammate into an unauthorized file deletion.
* **The Happy Path (Tasks):**
    1. Researcher teammate sends a mailbox request to the Coder teammate to "Delete .env".
    2. CMCS interceptor checks the Researcher's "Mesh Token."
    3. The token's role (Researcher) is compared against the authorized commands for that mission root.
    4. "Delete" is not an authorized command for the Researcher role.
    5. Request is blocked, and the mission root is notified of the violation.

## 4. Design & Architecture
* **System Flow:**
    * Every T2T mailbox request must include a CMCS Mesh Token in its headers.
    * The Mesh Token is generated during the initial "Mission Root" handshake.
    * The CMCS Interceptor validates the token signature and authorized role before forwarding to the recipient's mailbox.
* **APIs / Interfaces:**
    * `cmcs.v1.IssueMeshToken(mission_root_id, agent_id, role)`: Generates a session-bound token.
    * `cmcs.v1.ValidateMeshCommand(token, command)`: Checks if the command is authorized for the role.
* **Data Storage/State:**
    * Mission-bound role registries are stored in the core Blackboard (Shared KV Store).

## 5. Alternatives Considered
* **Parent-in-the-Loop:** Rejected as it creates a bottleneck for horizontal swarms.
* **Flat Identity Tokens:** Rejected as they do not distinguish between agent roles within a mission.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens must be session-bound and automatically revoked upon mission termination.
* **Observability:** Provides a "Command Lineage" view in the UI.

## 7. Evolutionary Changelog
* **2026-05-29:** Initial Document Creation.
