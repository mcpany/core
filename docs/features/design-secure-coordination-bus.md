# Design Doc: Secure Coordination Bus
**Status:** Draft
**Created:** 2026-05-16

## 1. Context and Scope
With the shift toward parallel agent swarms (e.g., Claude Code Agent Teams), the primary attack surface has moved from individual tool calls to the communication channels between teammates. Malicious or compromised agents can perform "Identity Shadowing" (spoofing instructions) or "Mailbox Injection" (poisoning the task list) to hijack the swarm's collective reasoning.

The Secure Coordination Bus provides a cryptographically hardened transport for inter-teammate messaging and state reconciliation. It ensures that every message within the swarm is authenticated, immutable, and bound to the session's mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce mandatory cryptographic signing for all teammate-to-teammate messages.
    * Provide a non-repudiable audit trail for all "Snapshot-and-Merge" state reconciliation events.
    * Implement sub-millisecond message validation to support high-frequency swarm coordination.
    * Bind every coordination message to a hardware-attested Mission Token.
* **Non-Goals:**
    * Managing the LLM-to-user communication (handled by the A2UI Gateway).
    * Providing long-term archival of coordination messages beyond the mission lifecycle.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars or allowing instruction spoofing.
* **The Happy Path (Tasks):**
    1. The Lead Agent initializes the Secure Coordination Bus with a Mission Token.
    2. Teammate agents (e.g., Coder and Reviewer) join the bus, providing their hardware-attested identity tokens.
    3. The Coder agent sends a "Review Request" message. The Bus signs the message with the Coder's session key.
    4. The Reviewer agent receives the message and verifies the signature and mission-root alignment.
    5. The Hub logs the authenticated handoff to the session audit trail.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A1[Agent 1] -- Signed Message --> Bus[Secure Coordination Bus]
        Bus -- Verify Signature/Intent --> Bus
        Bus -- Relayer --> A2[Agent 2]
        Bus -- Append --> Audit[Audit Trail]
    ```
* **APIs / Interfaces:**
    * `SendMessage(payload, signature, mission_token)`: Sends an authenticated message to the bus.
    * `ReceiveMessage(subscriber_id)`: Securely retrieves messages intended for the agent.
    * `ReconcileState(delta_set, proof)`: Performs signed state merging for the Blackboard.
* **Data Storage/State:** Uses memory-mapped buffers for high-speed message passing and an append-only, signed log for the coordination history.

## 5. Alternatives Considered
* **Plaintext Message Bus (e.g., Redis Pub/Sub)**: Rejected due to lack of identity verification and susceptibility to local port hijacking.
* **Blockchain-Based Coordination**: Rejected due to prohibitive latency for local swarm execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements "Auth-at-the-Pipe" for all connections. Every message is subject to a "Reasoning-Aware" check to ensure it doesn't leak context shards.
* **Observability:** Integration with the "Parallel Intent Visualizer" to show cryptographically verified message flows.

## 7. Evolutionary Changelog
* **2026-05-16:** Initial Document Creation.
