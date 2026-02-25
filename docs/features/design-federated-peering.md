# Design Doc: Federated MCP Node Peering

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agent toolsets expand beyond single-host environments, there is an increasing need to share and discover tools across distributed networks. The "Global Tool Mesh" represents a future where agents can "borrow" capabilities from remote nodes. MCP Any needs a mechanism to securely peer with other MCP Any instances (or compatible nodes) to create a federated tool registry while maintaining strict local Zero-Trust security boundaries.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable secure discovery of tools across multiple MCP Any nodes.
    *   Implement "Zero-Trust Proxying" where remote tools are subject to local policy firewall rules.
    *   Support cryptographic attestation of remote tool origins.
    *   Maintain low-latency discovery via similarity-based "Lazy Loading" of remote tool schemas.
*   **Non-Goals:**
    *   Creating a centralized, global tool directory (the system should be decentralized/peer-to-peer).
    *   Automatically executing remote code locally (execution remains on the remote node).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Allow a local developer's agent to use a proprietary "Internal Data Search" tool hosted on a central company MCP node, without the developer having to configure the tool manually.
*   **The Happy Path (Tasks):**
    1.  Architect configures a "Peer Relationship" between the local MCP Any instance and the Company Core MCP node.
    2.  The developer's agent requests a tool for "searching internal data."
    3.  Local MCP Any instance queries the Company Core peer via the `Federated Discovery Protocol`.
    4.  The Company Core node returns a signed tool schema.
    5.  Local MCP Any instance validates the signature and checks the local `Policy Firewall`.
    6.  The tool is exposed to the developer's agent; execution is proxied securely to the Company Core node.

## 4. Design & Architecture
*   **System Flow:**
    - **Peer Discovery**: Nodes use a gossip-like protocol or pre-configured static peers to exchange health and capability summaries.
    - **Discovery Proxy**: The `Lazy-MCP` middleware is extended to query remote peers when a local tool search miss occurs.
    - **Secure Proxying**: Tool calls are encapsulated in mTLS-encrypted tunnels, with `Recursive Context` headers preserved across hops.
*   **APIs / Interfaces:**
    - **Peering API**: `/v1/federation/peer` for handshake and capability exchange.
    - **Remote Tool Execution**: JSON-RPC over mTLS/HTTP2 for low-latency tool invocation.
*   **Data Storage/State:** Peer registry and remote tool metadata are stored in the local `Service Registry` with TTL-based eviction.

## 5. Alternatives Considered
*   **Centralized Registry (e.g., NPM for Tools)**: *Rejected* due to security concerns and the "Single Point of Failure" risk.
*   **VPN/Tunneling at Network Layer**: *Rejected* because it doesn't provide the granular, tool-level visibility and policy enforcement required for Zero-Trust Agency.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Remote tools are treated as "Untrusted" until their provenance is verified. All remote calls are subject to local `Intent-Aware` policy checks.
*   **Observability:** Federated calls are tracked with a global `Trace ID`, showing the full path across nodes in the `Agent Chain Tracer`.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
