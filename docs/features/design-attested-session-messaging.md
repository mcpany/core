# Design Doc: Attested Session Messaging (ASM)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agent swarms evolve from simple linear chains to complex, parallelized networks (like OpenClaw), the need for a low-latency, secure communication bus becomes critical. Standard MCP tool calls are too heavy for high-frequency "heartbeat" or "status" signals between agents. ASM provides a dedicated, authenticated messaging layer that allows agents to exchange state and control signals without the overhead of a full tool execution loop.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a low-latency messaging bus for inter-agent communication.
    * Ensure all messages are cryptographically "Attested" (linked to a valid session and agent identity).
    * Support "Direct Session Messaging" for deterministic delegation.
    * Integration with the Recursive Context Protocol to maintain lineage.
* **Non-Goals:**
    * Replacing the main MCP tool call mechanism for heavy tasks.
    * Providing long-term persistent storage (use the Shared KV Store instead).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent System Orchestrator.
* **Primary Goal:** Coordinate a swarm of 5 "Micro-Agents" working on a parallel data processing task, receiving real-time progress updates.
* **The Happy Path (Tasks):**
    1. Parent agent spawns 5 Micro-Agents using the `Recursive Context Protocol`.
    2. Each Micro-Agent registers with the ASM bus using its session token.
    3. Micro-Agents send periodic "progress" messages via ASM: `asm.send(parent_id, {progress: 0.2})`.
    4. Parent agent listens for these signals and updates its internal state or the UI in real-time.
    5. ASM verifies the "Attestation" of each message to ensure it's not from a rogue/injected process.

## 4. Design & Architecture
* **System Flow:**
    * **Handshake:** Agents authenticate with MCP Any using their session capability token.
    * **Messaging:** A pub/sub or direct-addressing model implemented over WebSockets or gRPC streams.
    * **Attestation:** Every message contains a signature or JWT that proves the sender is the authorized agent for that session.
* **APIs / Interfaces:**
    * `asm.send(recipient_id, payload)`
    * `asm.subscribe(topic, callback)`
    * `asm.broadcast(payload)`
* **Data Storage/State:** Transient state stored in-memory; lineage and audit logs stored in the main Audit Log.

## 5. Alternatives Considered
* **Polling the Shared KV Store:** agents poll a database for updates. *Rejected* due to high latency and database contention.
* **Custom MCP Notification Methods:** using standard MCP notifications. *Rejected* as many clients don't support custom notifications and they lack cross-agent routing.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ASM uses the same "Policy Firewall" logic. Agents can only message other agents within their "Intent-Scope."
* **Observability:** ASM messages are visible in the "Agent Chain Tracer" UI, allowing developers to debug swarm communication.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
