# Design Doc: Agentic Mesh DNS (AM-DNS) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve from single-node execution to distributed meshes (e.g., OpenClaw's Sovereign Node Tunneling), the reliance on static IP addresses and port numbers for inter-agent communication has become a major bottleneck and security risk. Mobile devices, cloud containers, and dynamic local networks make IP-based addressing brittle and difficult to attest.

The Agentic Mesh DNS (AM-DNS) Provider in MCP Any aims to solve this by providing a decentralized, cryptographically signed naming service. It allows agents to discover and address their peers using stable, mission-bound handles (e.g., `specialist.mission-root.mcp`) regardless of their physical network location. This ensures trust continuity as agents migrate across nodes.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a stable, human-readable (or agent-readable) naming convention for distributed agent nodes.
    * Implement cryptographic signing for all name-to-address mappings to prevent "Name Spoofing."
    * Support dynamic re-binding of handles as agents migrate between local, cloud, and mobile environments.
    * Integrate with Namespace-Locked Discovery (NLD) to ensure capability visibility is bound to the resolved identity.
* **Non-Goals:**
    * Replace traditional public DNS for non-agent traffic.
    * Provide a general-purpose P2P file sharing network.
    * Manage the underlying network tunnels (handled by the AMT Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Securely delegate a data-analysis task to a specialist subagent running on a separate edge device without manual IP configuration.
* **The Happy Path (Tasks):**
    1. The Mission Root agent spawns a "Data Specialist" on a remote node via MCP Any.
    2. The remote node registers its dynamic address with the AM-DNS Provider using the mission-root token.
    3. The Mission Root agent requests the capability `analyze_logs` from `specialist.mission-root.mcp`.
    4. AM-DNS resolves the handle to the current hardware-attested P2P tunnel endpoint.
    5. The task is delegated over an encrypted tunnel without the Mission Root ever needing to know the specialist's IP.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [AM-DNS Resolve Request] -> [MCP Any AM-DNS Hub] -> [Attested Address Map] -> [Resolved Endpoint]`
* **APIs / Interfaces:**
    * `rpc RegisterHandle(handle string, attestation_token string) returns (status)`
    * `rpc ResolveHandle(handle string) returns (endpoint_metadata)`
    * Integration with UACO for task card addressing: `target: "specialist.mission-root.mcp"`
* **Data Storage/State:**
    * Distributed KV store (Blackboard) backed by hardware-attested snapshots for the name registry.
    * TTL-based cache for resolved handles to minimize resolution latency.

## 5. Alternatives Considered
* **Static IP Configuration:** Rejected due to brittleness in dynamic environments (DHCP, NAT, Roaming).
* **mDNS/Bonjour:** Rejected because it lacks mission-bound cryptographic attestation and doesn't scale across WAN/Multi-cloud boundaries easily.
* **Centralized Registry:** Rejected to maintain the "Sovereign Node" philosophy; AM-DNS will use a federated gossip protocol between MCP Any nodes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All registrations must be signed by the Mission Root or a hardware-attested delegate. Resolution requests are audited against the session's NLD policy.
* **Observability:** Resolution latency and "Name Collision" events are logged to the Mesh Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
