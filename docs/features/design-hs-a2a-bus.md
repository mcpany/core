# Design Doc: High-Speed Inter-Agent Bus (HS-A2A)
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
Modern agentic swarms (e.g., OpenClaw, CrewAI) require high-frequency, low-latency communication to synchronize state, share research findings, and delegate sub-tasks at "machine speed." Current MCP transport mechanisms (Stdio, HTTP/JSON-RPC) introduce significant serialization and network overhead, which becomes a bottleneck for swarms of 10+ agents operating on the same host. The High-Speed Inter-Agent Bus (HS-A2A) aims to provide a shared-memory transport layer to eliminate this overhead.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a sub-millisecond latency transport for inter-agent communication on a single host.
    * Support shared-memory primitives for large data transfers (e.g., passing 10MB of research context without copying).
    * Maintain compatibility with existing MCP JSON-RPC schemas.
    * Implement strictly isolated "Bus Channels" governed by the Policy Engine.
* **Non-Goals:**
    * Replacing network-based A2A for distributed agents (across different hosts).
    * Implementing a custom agent framework (we are the bus, not the agent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator running a multi-agent refinement loop.
* **Primary Goal:** Share a massive context window between an "Architect" agent and 5 "Worker" agents without hitting I/O bottlenecks.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a `SwarmSession` via MCP Any.
    2. MCP Any allocates a shared memory segment for the session.
    3. Architect agent writes state to the segment and signals the Bus.
    4. Worker agents receive an interrupt/signal and read the shared state directly.
    5. State updates are atomic and governed by a lock manager.

## 4. Design & Architecture
* **System Flow:**
    `Agent A` -> `Shared Memory (mmap/shm)` <- `Agent B`
    `Control Plane`: `MCP Any` manages segment allocation, signaling (Unix sockets or named pipes), and lifecycle.
* **APIs / Interfaces:**
    * `mcp_a2a_create_channel(session_id, capacity)`
    * `mcp_a2a_publish(channel_id, data_ptr)`
    * `mcp_a2a_subscribe(channel_id, callback)`
* **Data Storage/State:**
    * Metadata stored in `sessions.db`.
    * Payload stored in `POSIX shared memory` (/dev/shm).

## 5. Alternatives Considered
* **gRPC over Unix Domain Sockets**: Faster than HTTP, but still involves serialization and copying. Rejected for high-frequency swarm needs.
* **Redis Pub/Sub**: Good for distributed systems, but adds an external dependency and network hop. Rejected for local-first performance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Access to shared memory segments is strictly bound to `Agent IDs` verified by the parent process. No raw memory access is allowed without a valid `Capability Token`.
* **Observability**: Throughput and latency metrics are exposed via a new `A2A Metrics` dashboard.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
