# Design Doc: Gossip-based Federated Peering

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As MCP Any scales to support large-scale agent deployments, the current model of manually configuring every federated node ("Centralized Configuration") has become a bottleneck. Federated MCP nodes need a way to autonomously discover each other, negotiate peering agreements, and share tool registries without a central point of failure or manual overhead.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a decentralized discovery protocol for MCP Any nodes using Gossip (SWIM/Serf).
    *   Enable automatic synchronization of tool registries across peered nodes.
    *   Provide a "Zero-Trust" negotiation phase before any tool access is granted.
    *   Ensure the system scales to 100+ distributed nodes with minimal configuration.
*   **Non-Goals:**
    *   Building a general-purpose P2P file sharing network.
    *   Implementing agent execution on remote nodes (peering only handles discovery/proxying).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Platform Engineer at a large enterprise.
*   **Primary Goal:** Deploy 20 new MCP Any nodes across different regions and have them automatically discover each other's tools.
*   **The Happy Path (Tasks):**
    1.  Engineer deploys 20 instances of MCP Any.
    2.  Each instance is provided with a list of 1-2 "Seed Nodes."
    3.  Instances use the Gossip protocol to discover all other 19 nodes.
    4.  Nodes exchange cryptographically signed "Manifests" of available tools.
    5.  An agent connected to Node A can now see and call tools hosted on Node T in a different region, with discovery happening automatically.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Layer**: Uses UDP Gossip (SWIM) for membership and health checking.
    - **Manifest Exchange**: Uses a secure TCP side-channel to exchange tool registries (Lazy-loaded).
    - **Proxy Layer**: Standard MCP tool calls are proxied over mTLS-encrypted connections between nodes.
*   **APIs / Interfaces:**
    - `X-MCP-Peer-ID`: Header for identifying the originating node.
    - Internal Gossip Port: 7946 (Default).
*   **Data Storage/State:** Peer registry stored in memory and periodically persisted to the `Shared KV Store`.

## 5. Alternatives Considered
*   **Centralized Consul/Etcd**: Rejected because it introduces a central point of failure and increases infrastructure complexity for smaller deployments.
*   **Static Peering**: Rejected because it doesn't scale and is prone to configuration drift.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All peering requests must be signed by a trusted Certificate Authority. "Gossip Poisoning" is mitigated by validating all incoming manifests against the Policy Firewall.
*   **Observability:** A "Mesh View" in the UI showing node health, latency between peers, and tool distribution.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
