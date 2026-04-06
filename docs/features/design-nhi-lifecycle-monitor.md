# Design Doc: NHI Lifecycle Monitor (NLM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The rapid proliferation of AI agents has led to an explosion of Non-Human Identities (NHIs). Many of these identities are task-specific and short-lived, yet their tokens often remain valid long after the mission is complete. These "orphaned" or abandoned accounts represent a massive attack surface for credential theft and lateral movement.

The NHI Lifecycle Monitor (NLM) provides active governance over agent identities. It tracks the lineage, task-binding, and real-time activity of every agent token, automatically revoking access when missions are terminated or when inactivity thresholds are met.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor the real-time activity and mission-status of all agent identities (NHIs).
    * Automatically revoke tokens for orphaned or abandoned agent sessions.
    * Enforce "Just-in-Time" (JIT) privilege pruning as missions evolve.
    * Provide a centralized dashboard for NHI inventory and risk assessment.
* **Non-Goals:**
    * Replacing core identity providers (e.g., OIDC/SAML). NLM manages the *lifecycle* of the tokens issued.
    * Managing human user identities.

## 3. Critical User Journey (CUJ)
* **User Persona:** Platform Security Engineer
* **Primary Goal:** Prevent an abandoned subagent token from being used by an attacker to access internal databases.
* **The Happy Path (Tasks):**
    1. A specialist subagent is spawned for a one-hour data migration task.
    2. NLM registers the new identity, binding it to the specific mission-root and task ID.
    3. The subagent completes the task and the parent agent signals mission termination.
    4. NLM detects the termination signal and immediately broadcasts a revocation request to all connected tool gateways.
    5. A rogue process attempts to use the subagent's token 5 minutes later.
    6. The tool gateway rejects the request as the token is already revoked in the NLM registry.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Mission Root] --> B[NLM Registry]
        B --> C[Identity Lifecycle Manager]
        C --> D[Revocation Broadcaster]
        D --> E[Tool Gateways]
        F[Agent Activity Monitor] --> C
    ```
* **APIs / Interfaces:**
    * `nlm.RegisterIdentity(metadata, taskID) -> NHIToken`: Records a new agent identity.
    * `nlm.RevokeIdentity(tokenID) -> bool`: Forcefully expires a token.
    * `nlm.Heartbeat(tokenID)`: Extends token life based on active reasoning.
* **Data Storage/State:**
    * **NHI Registry:** Persistent SQLite store tracking identity metadata, parentage, and task state.

## 5. Alternatives Considered
* **Short-lived JWTs only:** Rejected because they don't allow for immediate revocation if a mission ends earlier than expected.
* **Manual Revocation:** Rejected as it cannot keep pace with the machine-speed spawn rates of agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** NLM itself must be hardware-attested to prevent unauthorized registry modification.
* **Observability:** Integrated with the "NHI Lifecycle Dashboard" for real-time visualization of identity states.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
