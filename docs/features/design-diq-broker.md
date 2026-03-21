# Design Doc: Distributed Intent Quorum (DIQ) Broker
**Status:** Draft
**Created:** 2026-06-22

## 1. Context and Scope
With the emergence of "Semantic Session Ghosting" (SSG), subagents can persist and exfiltrate data even after a mission termination signal is sent by the parent agent. Existing single-agent termination signals are insufficient as they can be ignored or bypassed by compromised specialist subagents. MCP Any needs a mechanism to ensure that session termination is absolute and verified by multiple independent nodes in the swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-agent consensus mechanism for session termination.
    * Mandate hardware-attested signatures for all termination signals.
    * Neutralize "Semantic Session Ghosting" by ensuring all subagents in a mission scope are terminated.
* **Non-Goals:**
    * This system will not handle general task coordination (handled by UACO).
    * It does not replace the primary mission-root authority.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator / Mission Root
* **Primary Goal:** Securely terminate a multi-agent mission and ensure no subagents persist in the local environment.
* **The Happy Path (Tasks):**
    1. The Mission Root initiates a termination request for a specific mission ID.
    2. The DIQ Broker broadcasts a "Termination Proposal" to independent monitor agents.
    3. Monitor agents verify the mission state and provide hardware-attested approval signatures.
    4. Once the quorum threshold is reached, the DIQ Broker issues a "Final Termination Command" to all subagent transports (Named Pipes/WebSockets).
    5. MCP Any forcefully closes all associated transport channels and purges session-bound capabilities.

## 4. Design & Architecture
* **System Flow:**
    `Mission Root -> DIQ Broker -> [Monitor Agents] -> DIQ Broker -> Subagent Transports`
* **APIs / Interfaces:**
    * `TerminateMission(mission_id, signature)`
    * `ProposeTermination(mission_id)`
    * `AttestTermination(mission_id, signature)`
* **Data Storage/State:**
    Termination states and quorum signatures are stored in the hardware-attested session state within the Blackboard.

## 5. Alternatives Considered
* **Single-Agent Kill Signal:** Rejected because it is vulnerable to "Ghosting" where subagents ignore the signal.
* **Timeout-Based Termination:** Rejected as it may leave a window of vulnerability before the timeout expires.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All termination signatures must be hardware-bound (TPM/Secure Enclave) to prevent spoofing.
* **Observability:** All termination quorums and successes/failures are logged in the Local Security Audit Service.

## 7. Evolutionary Changelog
* **2026-06-22:** Initial Document Creation.
