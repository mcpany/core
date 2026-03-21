# Design Doc: Isolated Named-Pipe Transport Middleware
**Status:** Draft
**Created:** 2026-05-13

## 1. Context and Scope
With the identification of local network port exposure as a catastrophic vulnerability for multi-agent swarms (GSA-2026-OPENCLAW-ROUTING), MCP Any must provide a port-free alternative for inter-agent communication. The Isolated Named-Pipe Transport Middleware transitions inter-agent and inter-teammate coordination from the network stack to the kernel and filesystem, utilizing UNIX domain sockets (named pipes) for absolute isolation.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate the requirement for listening on local TCP/UDP ports for inter-agent communication.
    * Provide a high-performance (zero-copy where possible) transport layer via UNIX domain sockets.
    * Implement OS-level access control for the transport channel using standard filesystem permissions.
    * Integrate with the "Auth-at-the-Pipe" security model to verify agent identity at the transport layer.
* **Non-Goals:**
    * Providing remote inter-agent communication (this remains the role of the A2A Messaging Hub over TLS).
    * Replacing the MCP JSON-RPC protocol itself; this is a transport-level change.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars or network ports.
* **The Happy Path (Tasks):**
    1. The supervisor agent requests a "Secure Coordination Pipe" from MCP Any.
    2. MCP Any creates a UNIX domain socket at a protected project-local path (e.g., `.mcp/coord.sock`).
    3. MCP Any restricts the socket's permissions to the specific UID/GID of the authorized subagent team.
    4. Subagents connect to the pipe, providing a hardware-attested token for transport-level authentication.
    5. Inter-agent messages are exchanged entirely within the kernel, invisible to any network-level monitoring.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A1[Agent 1] -->|UNIX Socket| Hub[Named-Pipe Transport Hub]
        A2[Agent 2] -->|UNIX Socket| Hub
        Hub -->|OS-ACLs| FS[.mcp/coord.sock]
        Hub -->|Audit| eBPF[eBPF Socket Sentinel]
    ```
* **APIs / Interfaces:**
    * `CreatePipe(mission_id, uid_list)`: Provision a new isolated coordination socket with specific OS-level ACLs.
    * `SO_PEERCRED`: Kernel-level identity validation of the connecting process.
* **Data Storage/State:** Persistent socket handles in memory; filesystem nodes for the pipes.

## 5. Alternatives Considered
* **Local HTTP Tunneling:** Rejected because it remains vulnerable to port-hijacking and MitM attacks in multi-tenant or browser-connected environments.
* **Shared Memory (Zero-Copy):** Considered for performance, but rejected as the primary transport due to complexity in cross-process synchronization and lack of native kernel-level auditability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** We are adopting the "Auth-at-the-Pipe" model. All connections require hardware-attested identity tokens. The eBPF Socket Sentinel provides real-time semantic auditing to detect "Pipe-Siphoning."
* **Observability:** Kernel-resident trace scrubbing and eBPF logging compensate for the lack of network visibility.

## 7. Evolutionary Changelog
* **2026-05-13:** Initial Document Creation.
* **Update: 2026-05-13 - Integrating eBPF Socket Monitoring**
    **Context:** The move to port-free transport creates a visibility gap for traditional network monitoring tools.
    **Architecture Adjustment:** Mandatory integration with the eBPF Socket Sentinel for all named-pipe connections.
    **Security Impact:** Restores observability to inter-agent communication, enabling detection of unauthorized socket access.
