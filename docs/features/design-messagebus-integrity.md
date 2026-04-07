# Design Doc: Hardened Internal MessageBus (HIMB)
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
As agent infrastructure evolves from linear API sessions to complex subagent meshes (e.g., Gemini CLI Phase 3), the security of internal communication becomes paramount. Currently, inter-agent messages often lack the same level of auditing and authorization as external tool calls. If a single specialist subagent is compromised, it can use the internal MessageBus to coerce other subagents into unauthorized actions without triggering external security gates.

The Hardened Internal MessageBus (HIMB) provides a cryptographically signed, audited, and intent-bound transport for all internal agent communications, ensuring the integrity of the subagent mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mandatory, hardware-attested transport for all internal agent-to-agent messages.
    * Implement "Identity Injection" where every message is cryptographically bound to the sender's verified identity.
    * Enforce "Intent-Scoping" to ensure subagents can only send messages that align with the parent's verified mission root.
    * Maintain a hash-chained audit trail of all internal MessageBus activity.
* **Non-Goals:**
    * Managing external network traffic (handled by the Gateway/Proxy layers).
    * Providing a general-purpose message queue for non-agent applications.
    * Modifying the internal logic of connected LLMs.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a compromised "Code Reviewer" subagent from sending an unauthorized "Execute Command" message to a "Developer" subagent.
* **The Happy Path (Tasks):**
    1. The parent agent establishes a "Mission Root" on the HIMB.
    2. The compromised "Code Reviewer" subagent attempts to send a malicious message to the "Developer" subagent.
    3. The HIMB interceptor validates the message against the mission-root manifest and the "Code Reviewer's" hardware-attested identity.
    4. The HIMB identifies that the message exceeds the "Code Reviewer's" authorized intent scope.
    5. The message is blocked, and a "Sovereignty Breach" signal is sent to the parent agent and user dashboard.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent A] -->|Message| B[HIMB Hub]
        B --> C[Identity Injector]
        C --> D[Intent Validator]
        D -->|Verified| E[Subagent B]
        D -->|Blocked| F[Security Audit Log]
    ```
* **APIs / Interfaces:**
    * `himb.Publish(topic, payload, identityToken)`: Securely publishes a message to the bus.
    * `himb.Subscribe(topic, filter)`: Securely subscribes to authorized internal topics.
    * `himb.RotateKeys()`: Performs hardware-bound session key rotation.
* **Data Storage/State:**
    * Active mission intents are stored in the hardware-locked segment of the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Implicit Local Trust (Unauthenticated Pipes):** Rejected due to the risk of "Team Ghosting" and internal message spoofing.
* **External Message Brokers (RabbitMQ/Kafka):** Rejected for local agent environments due to excessive overhead and lack of TPM-binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory origin-locking and session-binding for all MessageBus listeners.
* **Observability:** Integrated with the "Agent Chain Tracer" for real-time visualization of internal message flow.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation (Iteration 2).
