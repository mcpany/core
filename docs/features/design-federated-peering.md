# Design Doc: Federated MCP Peering (Dynamic Discovery)

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
As MCP Any deployments scale across organizational boundaries, a centralized configuration model becomes a bottleneck and a single point of failure. Supplemental research shows that configuration propagation latency in large federated clusters can lead to tool discovery timeouts. We need a decentralized "Peering" model where MCP Any nodes can dynamically discover each other and share tool registries without a central registrar.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement decentralized node discovery via mDNS (local) and DHT (wide-area).
    *   Enable secure, authenticated peering between MCP Any instances.
    *   Support dynamic tool registry synchronization with eventual consistency.
    *   Reduce configuration propagation latency in 10+ node clusters.
*   **Non-Goals:**
    *   Global public tool discovery (peering is restricted to authorized nodes).
    *   Replacing the existing centralized model for small-scale deployments.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Platform Engineer (SRE).
*   **Primary Goal:** Scale the MCP infrastructure across three different data centers without manual configuration for every node.
*   **The Happy Path (Tasks):**
    1.  Engineer deploys a new MCP Any instance in Data Center B.
    2.  The instance uses the "Peering Protocol" to discover the existing mesh in Data Center A.
    3.  Nodes perform a mutual TLS handshake and exchange "Identity Attestations."
    4.  The new instance automatically receives a synced tool registry from its peers.
    5.  Agents in Data Center B can now call tools hosted in Data Center A via the local proxy.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Layer**: Uses mDNS for local network segments and a bootstrap-node DHT for cross-network peering.
    - **Peering Layer**: Persistent gRPC streams between nodes for real-time registry updates.
    - **Registry Sync**: Vector clocks ensure conflict-free merging of tool definitions across the mesh.
*   **APIs / Interfaces:**
    - `PeeringService`: Internal gRPC service for node-to-node communication.
    - `Discover()`: Method for finding peers.
    - `SyncRegistry()`: Method for exchanging tool definitions.
*   **Data Storage/State:** Mesh topology and remote tool metadata are stored in the local SQLite store, marked as `remote` source.

## 5. Alternatives Considered
*   **Centralized Configuration (Current)**: Rejected for high-scale environments due to propagation latency.
*   **Consul/Etcd Integration**: Rejected to minimize external dependencies and maintain the "Single Binary" philosophy of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All peering must use Mutual TLS with certificates pinned or verified against a shared CA. Peer nodes are subject to the same Policy Firewall as local tools.
*   **Observability:** A "Mesh Map" in the UI visualizes node health and synchronization latency.

## 7. Evolutionary Changelog
*   **2026-02-26:** Initial Document Creation.
