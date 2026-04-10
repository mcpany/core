# Design Doc: UMDP-Native Discovery Gateway
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve from local processes to distributed physical meshes, the "Discovery Phase" has become a major performance and security bottleneck. Agents currently spend up to 30% of their reasoning budget recursively probing nodes for capabilities, a phenomenon known as "Discovery Exhaustion."

MCP Any needs to implement the Universal Mesh Discovery Protocol (UMDP) to act as an authoritative discovery proxy. By standardizing how tools are broadcast and indexed across heterogeneous frameworks (OpenClaw, Claude Code, AutoGen), we can provide sub-millisecond capability mapping while maintaining zero-trust boundaries across physical nodes.

## 2. Goals & Non-Goals
* **Goals:**
    * Authoritative implementation of the UMDP v1.0 standard.
    * Framework-agnostic tool indexing and similarity-based discovery.
    * Hardware-attested tool registration to prevent "Shadow Capability" injection.
    * Integration with existing AMT (Attested Mesh Tunneling) for cross-node transport.
* **Non-Goals:**
    * Handling tool execution logic (delegated to standard MCP adapters).
    * Multi-cloud routing (handled by the global Mesh Resilience Hub).
    * User-facing UI for tool search (delegated to the Unified Discovery Manager).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Orchestrator
* **Primary Goal:** Discover a specific "Database Admin" tool located on a remote secure node without framework-specific configuration.
* **The Happy Path (Tasks):**
    1. Remote Node A broadcasts a UMDP capability beacon containing a cryptographically signed tool manifest.
    2. MCP Any UMDP Gateway receives the beacon and validates the hardware signature (TPM-bound).
    3. Local Agent B requests "Database Admin" capabilities via the local discovery bus.
    4. UMDP Gateway performs a similarity search and returns the masked capability card for the tool on Node A.
    5. Agent B completes a mission-bound handshake to unmask the full schema and establish a secure AMT tunnel.

## 4. Design & Architecture
* **System Flow:**
    [Remote Tool] --(UMDP Beacon)--> [UMDP Listener] --(Validation)--> [Global Capability Index]
    [Local Agent] --(Query)--> [Similarity Search Engine] --(Masked Card)--> [Local Agent]
    [Local Agent] --(Handshake)--> [UMDP Gateway] --(Unmask/Tunnel)--> [Remote Tool]

* **APIs / Interfaces:**
    * `GET /v1/discovery/mesh/search`: Similarity-based search for mesh-wide tools.
    * `POST /v1/discovery/mesh/register`: Ingests and validates UMDP capability beacons.
    * `x-umdp-node-id`: Header for identifying the originating physical node.

* **Data Storage/State:**
    * **Capability Index:** A persistent, encrypted SQLite store for tool metadata and node lineages.
    * **Attestation Cache:** Memory-mapped buffer for hardware-attested session keys to reduce handshake latency.

## 5. Alternatives Considered
* **Framework-Specific Bridges:** Building individual discovery bridges for OpenClaw, Claude, etc. Rejected due to "Bridge Fatigue" and the inability to handle future framework versions without constant updates.
* **Centralized Registry:** Maintaining a single global registry for all agents. Rejected due to the "Local Trust" mandate and the need for offline-first resilience in restricted networks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All beacons must be signed by an attested TPM. Capability cards are masked by default, ensuring that sensitive tool schemas are only revealed to peers within a verified mission scope.
* **Observability:** Discovery latency and "Search Hit/Miss" rates will be exported to the unified telemetry sink. Every discovery event will be linked to a mission-root trace ID.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
