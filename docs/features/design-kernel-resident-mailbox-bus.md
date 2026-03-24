# Design Doc: Kernel-Resident Mailbox Bus
**Status:** Draft
**Created:** 2026-05-17

## 1. Context and Scope
As agent swarms evolve into "Dynamic Teams" (Claude Code), inter-agent communication (Mailboxes) has become a primary target. High-trust loopback ports are vulnerable to hijacking and process impersonation. Attackers can inject malicious coordination messages into an agent's inbox, leading to "Mailbox Injection" and mission diversion.

The Kernel-Resident Mailbox Bus transitions inter-agent communication from high-level TCP/UDP loopback to isolated Unix Domain Sockets (UDS). By leveraging the OS kernel for peer-credential verification, we ensure that only verified teammate processes can exchange messages, effectively moving the security boundary to the kernel.

## 2. Goals & Non-Goals
* **Goals:**
    * transition inter-agent messaging to Unix Domain Sockets (UDS).
    * Implement kernel-level peer-cred verification (e.g., `SO_PEERCRED`).
    * Cryptographically sign all teammate-to-teammate messages.
    * Neutralize "Mailbox Injection" from unauthorized local processes.
* **Non-Goals:**
    * Will not replace A2A for remote (off-host) agent coordination.
    * Will not perform semantic scanning of messages (this is handled by the RIG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that a rogue script running on the user's machine cannot inject a "Terminate Mission" command into a teammate's inbox.
* **The Happy Path (Tasks):**
    1. Agent A and Agent B establish a coordination channel via the Mailbox Bus.
    2. The Bus creates a dedicated UDS file in a restricted `/run/mcp-any/` directory.
    3. Agent A sends a "Progress Report" to Agent B.
    4. The Mailbox Bus interceptor verifies Agent A's PID and UID via kernel peer-creds.
    5. The message is signed with Agent A's teammate-token and delivered to Agent B's inbox.
    6. A rogue script attempts to write to the UDS; the kernel rejects it due to UID/GID mismatch.

## 4. Design & Architecture
* **System Flow:**
    * Agent A -> UDS Write -> [Kernel Peer-Cred Check] -> Mailbox Bus Service -> Teammate Signature -> Agent B UDS Read.
* **APIs / Interfaces:**
    * `/var/run/mcpany/mailboxes/[agent_id].sock`: Dedicated UDS for each agent.
* **Data Storage/State:**
    * Transient socket files in memory-mapped filesystem (`tmpfs`).

## 5. Alternatives Considered
* **mTLS on Loopback:** Rejected due to the overhead of handshake latency and the complexity of managing local CA certs for ephemeral subagents.
* **Shared Memory (SHM):** Considered for performance but rejected as it lacks the built-in peer-cred verification features of UDS.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Relies on OS-level isolation and correct permissioning of the socket directory.
* **Observability:** Audit logs of failed peer-cred validations.

## 7. Evolutionary Changelog
* **2026-05-17:** Initial Document Creation.
