# Design Doc: Secure Agent Mesh (SAM)
**Status:** Draft
**Created:** 2026-05-31

## 1. Context and Scope
Today's "Universal Agent Infrastructure" relies on insecure inter-process communication (IPC) via local HTTP ports (e.g., localhost:8080). This architecture is vulnerable to port-scanning, interception, and cross-agent data leaks. In multi-agent swarms (CrewAI, AutoGen), a single compromised "subagent" could potentially access the host's environment variables or local files by hitting the common IPC port.

MCP Any must solve this by providing a Zero Trust "Secure Agent Mesh" that eliminates local port exposure in favor of authenticated, isolated communication channels.

## 2. Goals & Non-Goals
* **Goals:**
    * Replace HTTP/TCP localhost listeners with isolated **Docker-bound named pipes** or Unix Domain Sockets (UDS).
    * Enforce mutual authentication (mTLS or signed session tokens) for all inter-agent tool calls.
    * Provide a unified "Gateway" interface that abstracts the underlying IPC mechanism for agent developers.
* **Non-Goals:**
    * Encrypting on-disk agent logs (out of scope for the mesh).
    * Managing the lifecycle of agent containers (handled by K8s/Docker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share sensitive session context between 3 subagents (Manager, Coder, Reviewer) without exposing any local ports.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes MCP Any with a `mesh-id`.
    2. MCP Any creates a secure Unix Domain Socket (UDS) inside a shared Docker volume.
    3. Each subagent mounts the volume and authenticates with MCP Any using a transient session token.
    4. Subagent "Coder" calls the "Reviewer" via the `mcp-any://mesh/review` URI.
    5. MCP Any validates the request and routes it through the secure mesh, bypassing the host network stack.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A] <-> [Secure UDS / Named Pipe] <-> [MCP Any Gateway] <-> [Secure UDS / Named Pipe] <-> [Agent B]`
* **APIs / Interfaces:**
    * `RegisterSubagent(token, identity)`: Authenticates an agent joining the mesh.
    * `RouteCall(target_agent, payload)`: Securely routes a tool call within the mesh.
* **Data Storage/State:**
    * **State Manager:** Holds transient session keys in an in-memory, hardware-locked store.

## 5. Alternatives Considered
* **Local HTTP with API Keys:** Rejected due to the inherent risk of port-scanning on the host loopback interface.
* **gRPC over TLS:** Considered, but UDS/Named Pipes provide better isolation and lower latency for local-only communication.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All traffic is authenticated per call. The gateway enforces "Least Privilege" routing policies.
* **Observability:** All mesh traffic is logged with a common `mesh-id` correlation header, enabling swarm-wide tracing.

## 7. Evolutionary Changelog
* **2026-05-31:** Initial Document Creation.
