# Design Doc: Federated MCP Peering

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agentic workflows scale beyond single machines and departmental boundaries, the need for a distributed tool architecture arises. Federated MCP Peering allows multiple MCP Any instances to discover each other and securely share their connected tools. This transforms MCP Any from a standalone gateway into a global, decentralized tool mesh.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Secure, P2P discovery of remote MCP Any nodes.
    *   Transparent proxying of tool calls across federated nodes.
    *   Cross-node Zero-Trust policy enforcement.
    *   Resource-aware routing based on latency and cost telemetry.
*   **Non-Goals:**
    *   A centralized "global registry" (must remain decentralized).
    *   Automatic synchronization of secrets across nodes (each node manages its own secrets).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Platform Engineer.
*   **Primary Goal:** Allow an agent running in the "US-East" data center to use a legacy database tool that is only accessible via an MCP Any instance in "EU-West."
*   **The Happy Path (Tasks):**
    1.  Engineer configures a "Peering Relationship" between US and EU nodes via a mutual TLS handshake.
    2.  The US node automatically discovers the EU node's tool list: `eu_legacy_db_query`.
    3.  A US-based agent calls `eu_legacy_db_query(sql="...")`.
    4.  MCP Any US proxies the call to MCP Any EU, injecting the required `Agent-Attestation-Token`.
    5.  MCP Any EU validates the token and the Policy Firewall, executes the tool locally, and returns the result to the US node.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery**: Nodes use a gossip-based protocol or a shared DHT (Distributed Hash Table) for discovery.
    - **Proxying**: Tool calls are encapsulated in a secure gRPC tunnel between nodes.
    - **Policy Sync**: While nodes don't share secrets, they can subscribe to global Rego policies for consistent enforcement.
*   **APIs / Interfaces:**
    - **Peering API**: Internal `/v1/peer` endpoints for node-to-node handshakes.
    - **Unified Registry**: A virtual view of all local and peered tools.
*   **Data Storage/State:** Peering metadata is stored in the local KV store.

## 5. Alternatives Considered
*   **Centralized Tool Hub**: A single server that all agents connect to. *Rejected* due to latency and being a single point of failure.
*   **Manual VPN/Tunneling**: Requiring users to set up networking themselves. *Rejected* to reduce "Zero-Trust" configuration friction.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Peer-to-peer communication is encrypted via mTLS. Every proxied call must include the originating agent's identity for remote policy evaluation.
*   **Observability:** Federated traces must be stitched together to show the full multi-node execution path.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
