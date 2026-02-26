# Design Doc: Federated MCP Node Peering
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
As agentic workflows scale across organizations and infrastructure boundaries, a single centralized MCP server becomes a bottleneck and a single point of failure. Agents need to access tools distributed across multiple environments (local, cloud, edge). Federated MCP Node Peering allows multiple MCP Any instances to discover each other, share tool registries, and proxy tool calls securely across network boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Securely peer multiple MCP Any instances.
    * Synchronize tool discovery across a distributed mesh.
    * Implement latency-aware tool routing between nodes.
    * Maintain Zero-Trust boundaries (a node only shares tools it is authorized to share).
* **Non-Goals:**
    * Global public tool directory (peering is explicitly configured or discovery-bound).
    * Replacing standard MCP (it builds on top of it).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Cloud Agent Architect.
* **Primary Goal:** Allow an agent running in AWS to use a specialized local hardware tool (e.g., a local GPU-bound simulator) exposed via a local MCP Any instance.
* **The Happy Path (Tasks):**
    1. Admin configures peering between the AWS MCP Any instance and the Local MCP Any instance.
    2. The Local instance advertises its `gpu_simulator` tool to the AWS instance.
    3. The AWS agent sees `gpu_simulator` in its tool list (via the AWS instance).
    4. The agent calls the tool; the AWS instance proxies the call to the Local instance.
    5. The AWS instance injects telemetry (latency) to optimize future calls.

## 4. Design & Architecture
* **System Flow:**
    * **Handshake**: Nodes use mTLS and A2A-style attestation for peering.
    * **Registry Sync**: Incremental sync of tool schemas using a gossip-like protocol or a centralized coordinator (optional).
    * **Proxying**: gRPC-based tunneling of MCP JSON-RPC messages between nodes.
* **APIs / Interfaces:**
    * `/v1/peering/join`: Endpoint for nodes to initiate peering.
    * `/v1/federation/tools`: Stream of available tools across the mesh.
* **Data Storage/State:** Mesh topology and tool-to-node mapping stored in the `Shared KV Store`.

## 5. Alternatives Considered
* **Centralized Tool Registry**: A single database of all tools. *Rejected* due to latency and lack of local autonomy.
* **HTTP Tunneling (bore/ngrok)**: Using standard tunnels. *Rejected* because it doesn't provide the "Smart Routing" and security context needed for agentic workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All federated calls are subjected to the local Policy Firewall of *both* the source and destination nodes.
* **Observability:** Federated calls are tagged with `mesh_node_id` in traces to visualize cross-node latency.

## 7. Evolutionary Changelog
* **2026-02-26:** Initial Document Creation.
