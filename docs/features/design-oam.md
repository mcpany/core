# Design Doc: Optimistic Attestation Middleware (OAM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Sovereign Node Tunneling (SNT) and mandatory mesh encryption, inter-node tool calls now suffer from a significant "Tunneling Tax" (100ms-200ms) due to full cryptographic attestation at each hop. This latency breaks the "local-first" perception of the Universal Agent Bus.

MCP Any needs to decouple tool execution from the slow path of mission-root attestation. OAM provides a mechanism for speculative execution using temporary, hardware-bound session keys, allowing agents to proceed while full attestation completes asynchronously.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce perceived tool-call latency by up to 150ms in multi-node meshes.
    * Support speculative tool execution with automatic rollback on attestation failure.
    * Maintain Zero-Trust integrity by restricting speculative actions to a "Shadow Buffer."
* **Non-Goals:**
    * OAM will not bypass final attestation; it only reorders the execution vs. commitment.
    * OAM will not support non-idempotent tool calls that lack a native rollback/checkpoint mechanism.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Orchestrator
* **Primary Goal:** Execute a sequence of 10 cross-device tool calls in under 500ms despite mesh latency.
* **The Happy Path (Tasks):**
    1. Agent initiates a tool call to a remote node.
    2. OAM issues a "Speculative Trust Ticket" using an ephemeral hardware-bound key.
    3. Remote node executes the tool in a "Probabilistic Buffer."
    4. Full Mission-Root Attestation completes in the background.
    5. OAM verifies the attestation signature and commits the buffer to the global Blackboard.
    6. Agent receives the verified result with zero perceived handshake lag.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>OAM: Tool Call (Remote)
        OAM->>RemoteNode: Speculative Ticket + Encrypted Payload
        RemoteNode->>Sandbox: Execute Tool
        Sandbox-->>RemoteNode: Speculative Result
        RemoteNode-->>OAM: Encrypted Probabilistic Buffer
        OAM-->>Agent: Speculative Handover (Non-Blocking)
        Note over OAM: Background: Full Attestation
        Attestor->>OAM: Attestation Proof
        OAM->>Blackboard: Commit Result
    ```
* **APIs / Interfaces:**
    * `POST /v1/optimistic/execute`: Initiates a speculative call.
    * `GET /v1/optimistic/status/{ticket_id}`: Polls for final commitment status.
* **Data Storage/State:**
    * Speculative results are stored in an in-memory `memfd` buffer, isolated from the persistent Blackboard until commitment.

## 5. Alternatives Considered
* **Persistent Mesh Tunnels:** Rejected due to the high security risk of long-lived, un-attested tunnels in a Zero-Trust environment.
* **Aggressive Caching:** Rejected because tool results in dynamic swarms are highly context-dependent and cannot be reliably cached across missions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If background attestation fails, the speculative session is immediately revoked, and all state in the probabilistic buffer is purged. The calling agent is alerted to "Rewind" its reasoning.
* **Observability:** OAM logs include "Time-to-Optimistic" vs. "Time-to-Commit" metrics to monitor the effectiveness of latency reduction.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
