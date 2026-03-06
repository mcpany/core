# Design Doc: Gossip-Based Tool Discovery
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As MCP Any evolves from a single-node proxy to a Federated Tool Mesh, centralized tool registries (like a single `mcp_config.json` or a central server) become points of failure and bottlenecks. To support massive, distributed agent swarms, we need a decentralized way for nodes to discover and share tool capabilities.

This design introduces a P2P Gossip protocol that allows MCP Any instances to autonomously discover tools across network boundaries without a central coordinator.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable decentralized discovery of MCP tools across multiple federated nodes.
    * Provide eventual consistency for tool schemas across the mesh.
    * Minimize discovery latency through local caching of gossiped metadata.
    * Support health-aware routing (nodes gossip about which tools are currently reachable).
* **Non-Goals:**
    * Real-time global consensus (eventual consistency is sufficient).
    * Large file storage (gossip is for metadata/schemas only).
    * Replacing existing local configs (Gossip is an *additional* discovery provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** Federated Agent Mesh Architect
* **Primary Goal:** Automatically discover a "Postgres Query" tool hosted on Node B from an agent connected to Node A.
* **The Happy Path (Tasks):**
    1. Node B starts and loads its local Postgres MCP configuration.
    2. Node B joins the federated mesh and begins "gossiping" its tool metadata to known peers.
    3. Node A receives the gossip message and adds the Postgres tool schema to its local "Mesh Index."
    4. An agent on Node A asks for "Postgres tools."
    5. Node A identifies the tool in its Mesh Index and transparently proxies the request to Node B.

## 4. Design & Architecture
* **System Flow:**
    * Nodes maintain a list of active "Peers."
    * Every $T$ seconds, a node selects a random subset of peers and sends them a "Gossip Packet" containing a delta of its known tools.
    * Nodes use a Version Vector or Lamport Clock to resolve conflicts and ensure they only process newer metadata.
* **APIs / Interfaces:**
    * `GossipService`: Internal service handling P2P UDP/TCP communication.
    * `MeshDiscoveryProvider`: A new implementation of the `DiscoveryProvider` interface that reads from the local gossip cache.
* **Data Storage/State:**
    * In-memory cache for fast lookups.
    * Periodic persistence to a local "Mesh State" SQLite table for recovery after restart.

## 5. Alternatives Considered
* **Centralized Registry (etcd/Consul):** Rejected due to operational complexity and the desire for "Zero-Config" federation in small swarms.
* **Static Peer Lists:** Rejected as it doesn't scale well; Gossip allows for dynamic discovery of new nodes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All gossip packets must be signed with the node's private key. Peers will only accept gossip from nodes whose public keys are in their "Trusted Mesh" list.
* **Observability:** Track "Gossip Convergence Time" and "Mesh Health" metrics.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
