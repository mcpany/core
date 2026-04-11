# Design Doc: Active Intent Steerage (AIS) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agent swarms often suffer from "Reasoning Momentum," where they continue down a sub-optimal path despite changing user requirements or emerging environment feedback. Currently, correcting a running swarm often requires a full session reset, which is token-expensive and loses the existing "Blackboard" state.

The Active Intent Steerage (AIS) Broker provides a mechanism to securely inject "Corrective Intents" into an active reasoning loop. It ensures that instructions like "Stop refactoring and focus on the bug" are ingested immediately by all teammates without breaking the cryptographically signed chain of command or requiring a cold boot.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate real-time instruction injection into running agent sessions.
    * Maintain the integrity of the hardware-attested "Mission-Root" while allowing for "Intent Steering."
    * Propagate corrective intents across all connected teammates in a mesh simultaneously.
    * Integrate with the A2UI Gateway to provide a user interface for real-time steering.
* **Non-Goals:**
    * Replacing the primary mission-root (AIS is for corrective steering within an authorized mission).
    * Bypassing security guardrails (corrective intents are subject to the same Zero-Trust policy engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Pivot a 3-agent team from "Documentation Generation" to "Immediate Security Patching" when a high-severity CVE is discovered, without losing the context of the files they are currently analyzing.
* **The Happy Path (Tasks):**
    1. The user observes the swarm is spending too much time on docstrings.
    2. The user submits a "Corrective Intent" via the A2UI Dashboard: "Drop the docs, prioritize the CVE in auth.go."
    3. The AIS Broker validates the instruction against the user's hardware identity.
    4. The AIS Broker generates a "Steerage Token" and broadcasts it to the teammate mesh via the T2T Bridge.
    5. The agents receive the token, acknowledge the pivot in their internal monologues, and immediately shift their reasoning focus.
    6. The swarm applies the patch, and the mission continues.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[User/A2UI] -->|Corrective Intent| B[AIS Broker]
        B -->|Identity Validation| C[TPM/Secure Enclave]
        C -->|Signed Steerage Token| B
        B -->|Broadcast| D[T2T Bridge]
        D -->|Inject| E[Teammate A Context]
        D -->|Inject| F[Teammate B Context]
        E --> G[Reasoning Pivot]
        F --> H[Reasoning Pivot]
    ```
* **APIs / Interfaces:**
    * `POST /v1/steerage/inject`: Injects a corrective instruction into a specific mission scope.
    * `GET /v1/steerage/history`: Returns the timeline of steering events for a session.
* **Data Storage/State:**
    * Steerage tokens are appended to the "Reasoning Provenance" log in the Shared KV Store.

## 5. Alternatives Considered
* **Session Reset:** Rejected because it is too slow and loses non-persistent state.
* **Direct WebSocket Injection:** Rejected as it lacks the cryptographic provenance required for Zero-Trust swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Steerage tokens must be signed by the Mission-Root authority. Unauthorized subagents cannot use the AIS Broker to steer their siblings.
* **Observability:** The "Reasoning Alignment Monitor" visualizes how quickly agents align with the new steerage instructions.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
