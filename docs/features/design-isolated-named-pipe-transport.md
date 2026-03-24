# Design Doc: Isolated Named-Pipe Transport

**Status:** Draft
**Created:** 2026-05-13

## 1. Context and Scope
The industry pivot away from local network port exposure (GSA-2026-OPENCLAW-ROUTING) has made TCP/UDP loopback for inter-agent communication obsolete. "ClawdBot" style unauthenticated loopback hijacking has proven that any open port is a liability. This design doc outlines the transition to isolated named pipes (UNIX domain sockets) as the primary transport for all local inter-agent and inter-teammate coordination within MCP Any.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate all local network port usage for inter-agent communication.
    * Utilize OS-level filesystem permissions to restrict access to communication channels.
    * Provide a higher-performance transport than traditional loopback (lower latency, higher throughput).
    * Support "Auth-at-the-Pipe" identity validation.
* **Non-Goals:**
    * Replacing remote transport (mTLS/UACO) for agents running on different hosts.
    * Modifying the LLM messaging protocol (MCP/gRPC).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Architect
* **Primary Goal:** Securely connect a "Coder" subagent to a "Linter" subagent without exposing ports on the host.
* **The Happy Path (Tasks):**
    1. The supervisor agent requests a connection to the Linter subagent.
    2. MCP Any creates a secure, Docker-bound named pipe (e.g., `/run/mcp/linter.sock`).
    3. The supervisor agent presents a hardware-attested token to the socket.
    4. The kernel validates the calling process and grants access to the named pipe.
    5. The two agents communicate entirely within the kernel-resident memory.

## 4. Design & Architecture
* **System Flow:**
    * Agents no longer listen on `127.0.0.1:PORT`.
    * MCP Any acts as the "Socket Broker," managing the lifecycle of UNIX domain sockets in a restricted directory.
    * Standard filesystem permissions (ACLs) are applied to the socket files to ensure only authorized agent processes can read/write.
* **APIs / Interfaces:**
    * `GetSocketPath(agent_id)`: Returns the filesystem path to the isolated named pipe.
    * `AuthorizePipe(socket_fd, auth_token)`: Performs transport-level identity validation.
* **Data Storage/State:** Socket state is ephemeral; permissions are managed via OS-level syscalls.

## 5. Alternatives Considered
* **Shared Memory**: Rejected due to higher complexity in managing cross-process state and synchronization.
* **mTLS over Loopback**: Rejected because it still leaves the port open to connection-exhaustion and MitM attempts before the handshake is complete.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Mandatory "Auth-at-the-Pipe" ensures that even if a process can "see" the socket file, it cannot communicate without a valid mission token.
* **Observability**: Named-pipe usage and throughput are monitored via the Kernel-Resident Trace Scrubber.

## 7. Evolutionary Changelog
* **2026-05-13:** Initial Document Creation.

### Update: 2026-05-17 - Transport-Layer Session Binding (TLSB)
**Context:** Today's research has identified a new "Team Ghosting" exploit pattern where sibling agents in a parallel swarm hijack stale or un-authenticated named-pipe sessions to exfiltrate context.
**Architecture Adjustment:**
* Deprecating optional loopback listeners in Section 4.
* Introducing **Transport-Layer Session Binding (TLSB)** for all named-pipe and local transport channels.
* Mandatory cryptographic binding of every inter-agent connection to a unique, hardware-attested **Reasoning Session Token**.
**Security Impact:** Prevents unauthorized context access or capability reuse by rogue sibling agents, ensuring session isolation even within the same mission root.
