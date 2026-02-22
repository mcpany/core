# Design Doc: Isolated Inter-Agent Communication (Named Pipes)
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agent swarms grow in complexity, the need for secure, low-latency communication between the gateway and local agents becomes critical. Current local communication often relies on HTTP or Stdio. HTTP is vulnerable to local port sniffing (as seen in the "MCP-Port-Sniff" vulnerability), and Stdio can be brittle and hard to multiplex for multiple subagents. MCP Any needs a secure, isolated transport for inter-agent communication.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a Zero Trust transport using Unix Domain Sockets (Named Pipes on Windows).
    * Mitigate the risk of unauthorized host-level traffic interception.
    * Enable high-performance, concurrent communication for subagent swarms.
* **Non-Goals:**
    * Replacing existing HTTP/Stdio adapters (they will remain for legacy/remote support).
    * Providing remote network isolation (this is purely for local agent-to-gateway comms).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (e.g., developer running OpenClaw or CrewAI).
* **Primary Goal:** Share secure context between 3 agents without exposing local traffic to other processes on the host.
* **The Happy Path (Tasks):**
    1. Orchestrator starts MCP Any with a designated "socket directory".
    2. MCP Any creates a secure Unix Domain Socket (UDS) for each authorized agent.
    3. Agents connect to their unique UDS instead of an HTTP port.
    4. Tool calls are routed through the UDS, ensuring no other process can sniff the JSON-RPC traffic.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent A] <---> [UDS /tmp/mcp-a.sock] <---> [MCP Any Core] <---> [Upstream Service]`
* **APIs / Interfaces:**
    * New transport type: `socket`.
    * Config option: `socket_path` or `socket_dir`.
* **Data Storage/State:**
    * Socket permissions are strictly set to the user running MCP Any (0600).

## 5. Alternatives Considered
* **Local HTTP with TLS:** Rejected due to the complexity of managing local certificates and the fact that ports are still "visible" to the OS.
* **Enhanced Stdio:** Rejected because it doesn't solve the discovery and multiplexing problem for many subagents easily.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Socket file permissions act as the first layer of defense. Only processes with the same UID can access the socket.
* **Observability:** Trace IDs will be propagated through the socket transport to maintain visibility in the "Agent Activity Feed".

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
