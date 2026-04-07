# Design Doc: Federated Reputation Quorum
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
The "ClawHavoc" registry compromise proved that centralized "Skill Registries" are single points of failure. Even with behavioral profiling, malicious skills can use "Delayed Payloads" to bypass initial vetting. A decentralized, consensus-based approach is needed to determine tool safety.

The Federated Reputation Quorum allows MCP Any nodes to peer with each other and reach a collective consensus on a tool's safety based on real-time, distributed behavioral signals.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable MCP Any nodes to share tool reputation scores and behavioral alerts.
    * Reach collective consensus (Quorum) on tool "Trusted" status.
    * Implement "Consensus-Driven Scoping" where a tool's capabilities are restricted if its reputation drops.
    * Neutralize "Registry Poisoning" by relying on distributed attestation rather than a single database.
* **Non-Goals:**
    * Providing a general-purpose blockchain for all agent data.
    * Managing tool "quality" (it focuses strictly on security and safety).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Agent User
* **Primary Goal:** Safely install a new "Database Connector" tool that has not yet been audited by the central registry.
* **The Happy Path (Tasks):**
    1. User attempts to install the "Database Connector".
    2. Local MCP Any node checks its internal registry and finds no record.
    3. The node broadcasts a "Reputation Query" to its Federated Quorum peers.
    4. Peers respond with their local behavioral logs and risk scores for that tool's hash.
    5. Local node calculates a "Collective Trust Score."
    6. If the score exceeds the "Trust Quorum" threshold, the tool is installed with "Restricted" permissions.
    7. As more nodes report safe behavior, the tool's reputation increases, eventually reaching "Trusted" status.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Local Node] <-->|Reputation Sync| B[Peer Node 1]
        A <-->|Reputation Sync| C[Peer Node 2]
        A <-->|Reputation Sync| D[Peer Node 3]
        E[Tool Discovery] --> A
        A --> F{Quorum Decision}
        F -->|Safe| G[Load Tool]
        F -->|Unsafe| H[Quarantine]
    ```
* **APIs / Interfaces:**
    * `quorum.QueryReputation(toolHash) -> TrustScore`: Collects signals from peers.
    * `quorum.BroadcastAlert(toolHash, alertType)`: Shares high-risk behavioral anomalies with the mesh.
* **Data Storage/State:**
    * **Distributed Reputation Ledger:** A local, synchronized cache of peer-attested tool scores.

## 5. Alternatives Considered
* **Centralized Security Updates (like Antivirus)**: Rejected because it's too slow to respond to machine-speed swarm attacks and creates a central target for attackers.
* **Pure Local Sandbox**: Rejected because it can't detect "Delayed Payloads" that trigger after the initial sandbox period.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Peer nodes must be hardware-attested (SMI) before their reputation signals are accepted into the quorum calculation.
* **Observability:** Integrated with the "UAB Reputation Explorer" in the UI.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
