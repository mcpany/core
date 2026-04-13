# Design Doc: Monotonic Coordination Anchoring (MCA) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of Sovereign Node Tunneling (SNT) and distributed multi-agent meshes, inter-node coordination has become a primary target for replay attacks. Standard session-based tokens (JWTs) are susceptible to "Contextual Replay" during their validity window, where an attacker captures a valid coordination fragment and replays it to a remote node to "shadow" or hijack a mission-root context.

The Monotonic Coordination Anchoring (MCA) Provider is required to provide a hardware-bound temporal integrity layer that ensures every inter-node coordination fragment is unique, sequential, and non-reusable.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-bound (TPM) monotonic timestamps for all inter-node coordination messages.
    * Enforce strict sequential validation for incoming coordination fragments at the mesh gateway.
    * Neutralize "Contextual Replay" attacks by making coordination fragments mission-phase dependent.
    * Provide a cryptographic proof of temporal lineage for the "Chain of Command."
* **Non-Goals:**
    * Managing low-level network encryption (handled by the AMT Broker).
    * Providing a global wall-clock synchronization service (MCA focuses on monotonic sequences).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Securely coordinate a code-review task between a Laptop agent and a remote Build-Server agent without risk of coordination replay.
* **The Happy Path (Tasks):**
    1. The Laptop agent generates a "Review Request" coordination fragment.
    2. The MCA Provider intercepts the fragment and appends a TPM-signed monotonic counter and Mission-Phase ID.
    3. The fragment is sent through the Attested Mesh Tunnel (AMT) to the Build-Server.
    4. The Build-Server's MCA Provider verifies that the counter is strictly greater than the last received counter for that Mission-Phase.
    5. The Build-Server agent processes the review and returns a signed "Review Result" with its own monotonic increment.
    6. If an attacker attempts to replay the Laptop's "Review Request," the Build-Server's MCA Provider rejects it as a "Stale Sequence Index."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B[MCA Provider]
        B --> C[TPM Monotonic Counter]
        C --> D[Signed Fragment]
        D --> E[Mesh Transport]
        E --> F[Receiver MCA Validator]
        F --> G{Is Sequential?}
        G -->|Yes| H[Agent Execution]
        G -->|No| I[Block Replay]
    ```
* **APIs / Interfaces:**
    * `mca.AnchorFragment(fragment, missionID) -> AnchoredFragment`: Appends temporal metadata and signs.
    * `mca.ValidateSequence(anchoredFragment) -> bool`: Verifies monotonicity and signature.
* **Data Storage/State:**
    * **Sequence Registry:** Local persistent store of the last-seen counter per Mission-Phase/Node pair.

## 5. Alternatives Considered
* **Short-lived Session Tokens**: Rejected because even 1-minute tokens allow for high-frequency replay in machine-speed swarms.
* **Vector Clocks**: Considered for decentralized coordination, but TPM-bound monotonic counters provide stronger hardware-level guarantees for mission sovereignty.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MCA is a prerequisite for "Echo-Immune Coordination Fragments."
* **Observability:** Sequence gaps and replay attempts are visualized in the "Coordination Replay Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
