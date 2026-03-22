# Design Doc: Federated Policy Synchronizer
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
As MCP Any deployments scale across enterprise departments and multi-cloud environments, managing security policies (Allowed Origins, Tool Blacklists, Content-Addressable Config hashes) becomes a significant operational burden. Administrators need a way to ensure that security guardrails are consistent across all "Satellite" nodes without manual configuration of each instance.

The `Federated Policy Synchronizer` provides a secure, decentralized bus for synchronizing these policies across a fleet of MCP Any instances, anchored to a central "Policy Authority" or a peer-to-peer trust mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically propagate "Allowed Origin" lists and `CAC` hashes to all connected nodes.
    * Support "Conflict-Free Replicated Data Types" (CRDTs) to handle concurrent policy updates from different admins.
    * Implement hardware-attested policy signatures to ensureSatellite nodes only accept authentic updates.
    * Provide a "Pull-based" fallback for nodes behind restrictive firewalls.
* **Non-Goals:**
    * Synchronizing agent state or Blackboard data (handled by the `A2A Messaging Hub`).
    * Direct management of non-MCP Any infrastructure.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Center (SOC) Manager
* **Primary Goal:** Immediately block a malicious browser origin across 50 MCP Any instances globally.
* **The Happy Path (Tasks):**
    1. The SOC Manager adds the malicious origin to the `BlockedOrigins` list in the "Master" MCP Any node.
    2. The Master node signs the update using its TPM-bound private key.
    3. The Federated Synchronizer broadcasts the signed "Policy Delta" to all registered Satellite nodes via a secure gossip protocol.
    4. Satellite nodes verify the signature against the Master node's public key.
    5. The new origin block is applied instantly across the entire fleet without a restart.

## 4. Design & Architecture
* **System Flow:**
    `Admin UI` -> `Policy Registry (Master)` -> `Broadcaster (Gossip/PubSub)` -> `Satellite Receiver` -> `Local Policy Engine`
* **APIs / Interfaces:**
    * Sync Protocol: gRPC Stream or Secure WebSocket for delta propagation.
    * Peer Discovery: mDNS (Local) or Static Peer List (Cloud).
* **Data Storage/State:**
    * **Policy Store**: Local SQLite database on each node storing the latest version-checked policy state.

## 5. Alternatives Considered
* **Centralized Database**: Rejected to avoid a single point of failure and to support air-gapped/offline node operation.
* **GitOps (Config Repo)**: A viable alternative, but rejected as the *primary* mechanism due to the latency of Git sync cycles for emergency origin blocking.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All policy updates MUST be cryptographically signed. Nodes use a "Secure Boot" process to establish the initial trust root for the Policy Authority.
* **Observability**: The `Federated Governance Dashboard` will show the synchronization status (drift) of every node in the fleet.

## 7. Evolutionary Changelog
* **2026-03-22:** Initial Document Creation.
