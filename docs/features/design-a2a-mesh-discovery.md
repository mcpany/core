# Design Doc: A2A Mesh Discovery Protocol
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
In a multi-agent ecosystem, agents often operate in silos. For a true "swarm" to be effective, agents need to discover each other's capabilities dynamically across different nodes. The A2A Mesh Discovery Protocol aims to turn distributed MCP Any instances into a unified, searchable "capability fabric," allowing any agent to find the most suitable peer for a specific sub-task.

## 2. Goals & Non-Goals
* **Goals:**
    * Peer-to-peer discovery of agent identities and capabilities.
    * Similarity-based searching of tool and agent descriptions.
    * Support for "Capability Attestation" (ensuring a peer is who they say they are).
    * Decentralized architecture without a single point of failure.
* **Non-Goals:**
    * A global, public registry for all agents (focus is on private/enterprise meshes).
    * Real-time routing of agent messages (handled by the A2A Gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** Decentralized Swarm Orchestrator
* **Primary Goal:** Find a "Security Audit Specialist" agent available on the local network mesh.
* **The Happy Path (Tasks):**
    1. Agent A queries its local MCP Any node for a capability matching "code security audit."
    2. The local node broadcasts a discovery request to its peered MCP Any instances.
    3. Agent B (the specialist) on Node 2 matches the criteria.
    4. Node 2 returns Agent B's identity, endpoint, and attestation signature to Node 1.
    5. Agent A establishes a secure A2A session with Agent B.

## 4. Design & Architecture
* **System Flow:**
    MCP Any instances maintain a "Capability DHT" (Distributed Hash Table) or use a gossip protocol to synchronize a high-level index of available agents and tools.
* **APIs / Interfaces:**
    * `MCP_DISCOVER(query)`: Search the mesh for matching capabilities.
    * `MCP_ANNOUNCE(identity, capabilities)`: Register an agent with the local mesh.
* **Data Storage/State:**
    * A local "Mesh Cache" stores known peer capabilities, refreshed via gossip.

## 5. Alternatives Considered
* **Centralized Registry (Consul/Etcd)**: Rejected to avoid single points of failure and to support air-gapped or transient network environments.
* **Static Configuration**: Rejected as it cannot handle the dynamic "on/off" nature of agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Every announcement must be cryptographically signed. Peers must verify the "Chain of Trust" before sharing resources.
* **Observability**: Mesh topology map visualization in the UI.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
