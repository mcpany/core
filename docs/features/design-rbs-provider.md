# Design Doc: Reputation-Bound Scoping (RBS) Provider
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
With the emergence of autonomous swarms like OpenClaw's SWARM protocol, agents are now collaborating across framework boundaries without direct human supervision. This increases the risk of "Machine-Speed Swarm Attacks," where a compromised peer can quickly spread malicious instructions or exhaust resources across the mesh.

MCP Any needs a mechanism to scope agent capabilities not just by identity, but by real-time reputation. The RBS Provider will act as the authoritative gateway that revokes or restricts tool access based on an agent's verified track record and hardware-attested reputation score.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized reputation registry within MCP Any.
    * Enforce hardware-attested reputation thresholds for all tool calls.
    * Support dynamic capability revocation for agents whose reputation falls below a mission-defined floor.
    * Integrate with the SWARM protocol's contribution-based scoring logic.
* **Non-Goals:**
    * Building a global, multi-tenant reputation database (focus is on local/mission scope).
    * Defining the specific scoring algorithms for every possible agent action (reputation is ingested, not purely calculated).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (DevSecOps)
* **Primary Goal:** Prevent a low-reputation specialist agent from executing destructive shell commands in a production build pipeline.
* **The Happy Path (Tasks):**
    1. The Orchestrator defines a mission policy in MCP Any requiring a minimum reputation score of 80 for `run_shell_command`.
    2. A specialist agent from an external framework attempts to join the mission mesh.
    3. The RBS Provider validates the agent's hardware-attested reputation token (e.g., from OpenClaw SWARM).
    4. The agent's score is verified at 45.
    5. The agent is successfully discovered in the mesh but its `run_shell_command` capability is automatically masked by the RBS gateway.
    6. The agent attempts to call the tool; the RBS Provider interdicts the call and logs a reputation-bound violation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>RBS Provider: Handshake (Identity + Rep Token)
        RBS Provider->>Registry: Verify TPM Signature
        Registry-->>RBS Provider: Valid Signature
        RBS Provider->>Policy Engine: Check Thresholds(Mission_ID)
        Policy Engine-->>RBS Provider: Min_Rep=80
        RBS Provider->>Discovery Hub: Mask Capabilities(Agent_ID, Score=45)
        Discovery Hub-->>Agent: Filtered Toolset
    ```
* **APIs / Interfaces:**
    * `GET /v1/reputation/{agent_id}`: Retrieve current score and attestation.
    * `POST /v1/reputation/verify`: Validate a peer's reputation token.
    * `X-MCP-Reputation-Threshold`: New header for capability tokens.
* **Data Storage/State:**
    * Reputation scores are stored in a dedicated table within the Shared KV Store (Blackboard).
    * Audit logs for reputation-based interdictions are persisted for compliance reporting.

## 5. Alternatives Considered
* **Static Capability Tokens:** Rejected because they don't account for behavioral decay or "zero-day" agent compromises.
* **Purely Local Scoring:** Rejected because it doesn't allow for reputation to persist across framework handoffs (e.g., Claude to OpenClaw).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All reputation tokens must be hardware-attested (TPM) to prevent score spoofing.
* **Observability:** Real-time visualization of agent scores and revocation events in the Reputation-Bound Scoping Dashboard.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
