# Design Doc: Secure Inter-Agent Communication (Named Pipes)
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
Autonomous agents and subagent swarms (like OpenClaw) currently rely on local network ports (HTTP/WebSocket) for communication. This pattern is vulnerable to spoofing and unauthorized access from other local processes or rogue subagents. MCP Any needs to provide a host-isolated, secure communication primitive to enable trustworthy agent swarms.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a communication channel that does not expose network ports on the host.
    *   Ensure that only authorized containers/processes can access specific agent communication pipes.
    *   Support the standard MCP protocol over these pipes.
*   **Non-Goals:**
    *   Implementing a full service mesh.
    *   Supporting cross-host communication (focus is on local/single-node swarms).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Securely connect a "Planner" agent to three "Worker" agents without exposing any network ports on the Mac Mini/VPS.
*   **The Happy Path (Tasks):**
    1.  Orchestrator starts MCP Any with the `named-pipe` transport enabled.
    2.  MCP Any creates isolated FIFO pipes in a protected Docker volume.
    3.  Agents mount the volume and communicate via MCP-over-JSON-RPC through the pipes.
    4.  MCP Any Policy Engine verifies every tool call passing through the pipe.

## 4. Design & Architecture
*   **System Flow:**
    `[Agent A (Container)] <--> [/vols/pipes/agent-a.pipe] <--> [MCP Any (Gateway)] <--> [/vols/pipes/agent-b.pipe] <--> [Agent B (Container)]`
*   **APIs / Interfaces:**
    *   New Transport: `mcpany.transport.v1.NamedPipe`.
    *   Config: `path: "/tmp/mcpany/pipes/agent-worker-1"`.
*   **Data Storage/State:**
    *   Pipe metadata and access control lists (ACLs) stored in MCP Any's internal SQLite state.

## 5. Alternatives Considered
*   **Local HTTP (Status Quo):** Rejected due to the "OpenClaw Exploit" risk and lack of host-level isolation.
*   **Unix Domain Sockets:** Strong candidate, but Named Pipes (FIFOs) offer simpler stream-based semantics for some container runtimes and are easier to visualize/debug as files.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Use Docker mount permissions and Unix file ownership to restrict pipe access. All traffic is inspected by the MCP Any Policy Engine (Rego/CEL).
*   **Observability:** Each pipe has a dedicated log stream in the MCP Any UI, allowing real-time monitoring of inter-agent traffic.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
