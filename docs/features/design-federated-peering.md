# Design Doc: Federated MCP Node Peering

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agentic workflows scale across organizations, a single centralized MCP server becomes a bottleneck and a single point of failure. Teams often maintain their own local tools that shouldn't be fully exposed but need to be accessible by authorized agents in other departments. Federated MCP Node Peering allows multiple MCP Any instances to discover each other and securely share tool access across network boundaries.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable secure discovery and peering between distributed MCP Any nodes.
    *   Allow "Tool Proxying" where Node A can call a tool on Node B on behalf of an agent.
    *   Enforce Zero-Trust policies across federated boundaries.
    *   Inject latency telemetry to enable performance-aware tool selection.
*   **Non-Goals:**
    *   Creating a public, unauthenticated tool registry.
    *   Synchronizing the entire tool database across all nodes (lazy-proxying only).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise DevOps Engineer.
*   **Primary Goal:** Allow a "Finance Agent" running on the corporate cluster to use a "Local Ledger Tool" running on a secure local workstation in the Finance department.
*   **The Happy Path (Tasks):**
    1.  The Finance Department Node and the Corporate Node establish a peered connection using mutual TLS and pre-shared keys.
    2.  The Corporate Node performs a "Lazy Discovery" query to the Finance Node for ledger-related tools.
    3.  The Finance Agent requests a tool call to the "Ledger Tool" via the Corporate Node.
    4.  The Corporate Node proxies the request to the Finance Node.
    5.  The Finance Node executes the tool locally and returns the result through the peered bridge.

## 4. Design & Architecture
*   **System Flow:**
    - **Peering Handshake**: Nodes exchange identity certificates and establish a persistent gRPC or WebSocket tunnel.
    - **Cross-Node Discovery**: Uses the `Lazy-MCP` protocol to query remote registries without downloading all schemas.
    - **Request Proxying**: The `FederationGateway` middleware intercepts tool calls destined for remote nodes and wraps them in a secure RPC envelope.
*   **APIs / Interfaces:**
    - `POST /v1/federation/peer`: Initiate peering.
    - `GET /v1/federation/tools`: Query peered tool index.
    - `POST /v1/federation/call`: Proxy a tool execution.
*   **Data Storage/State:** Peering metadata (Node IDs, public keys, latency history) is stored in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Centralized Tool Hub**: A single global registry for all tools. *Rejected* due to security concerns regarding local tool exposure and high latency.
*   **VPN-based Access**: Forcing all nodes onto a single flat network. *Rejected* as it violates the principle of Zero-Trust isolation between departments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All cross-node traffic is encrypted. The `Policy Firewall` on the target node must explicitly allow the source node's identity to call specific tools.
*   **Observability:** Peering health and cross-node latency are monitored and displayed in the Federated Node Manager UI.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
