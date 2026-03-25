# Design Doc: Teammate-to-Teammate (T2T) Encryption Bridge
**Status:** Draft
**Created:** 2026-05-22

## 1. Context and Scope
The introduction of "Agent Teams" in Claude Code and similar "Mesh" orchestration patterns in OpenClaw has created a need for secure, cross-framework horizontal communication. Teammates within a team need to exchange mailbox messages and synchronize a Shared Task List. Currently, these mechanisms are often unencrypted or framework-specific, creating a risk of state injection or intent hijacking if one teammate is compromised.

The T2T Encryption Bridge provides a universal, secure bus for teammate-to-teammate communication. It allows agents from different frameworks to coordinate with cryptographic guarantees of integrity and privacy.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide end-to-end encryption for inter-agent mailbox messages.
    *   Ensure cryptographic integrity for the Shared Task List across multiple frameworks.
    *   Implement "Intent-Bound Validation": ensure messages align with the verified Mission Root.
    *   Support framework-agnostic handshakes (Claude Code teammate <-> OpenClaw specialist).
*   **Non-Goals:**
    *   Replacing the agent frameworks' internal reasoning logic.
    *   Providing a public messaging service (T2T is scoped to a specific mission/session).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer orchestrating a "Heterogeneous Swarm" (Claude Code lead + 2 OpenClaw subagents).
*   **Primary Goal:** Securely delegate a database migration task from the Claude lead to an OpenClaw specialist and monitor its progress via the Shared Task List.
*   **The Happy Path (Tasks):**
    1.  The User initializes an Agent Team via MCP Any.
    2.  MCP Any establishes a T2T Encryption Bus for the session.
    3.  The Claude Code lead teammate posts a "Database Migration" task to the Shared Task List.
    4.  The OpenClaw specialist teammate sees the task and "claims" it.
    5.  The OpenClaw agent sends a direct mailbox message to the Claude lead requesting the schema.
    6.  The T2T Bridge encrypts the message using the lead's public key.
    7.  The Claude lead receives and decrypts the message, then responds with the schema.
    8.  The T2T Bridge validates that the exchange is within the "Mission Root" intent scope.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        AgentA[Claude Teammate] <--> Bridge[T2T Encryption Bridge]
        AgentB[OpenClaw Teammate] <--> Bridge
        Bridge <--> STL[Shared Task List (SQLite/Encrypted)]
        Bridge <--> Mailbox[Inter-Agent Mailbox (Encrypted)]

        subgraph "Security Layer"
            Bridge --> VAL[Mailbox Integrity Middleware]
            VAL --> POL[Mission-Root Policy]
        end
    ```
*   **APIs / Interfaces:**
    *   `POST /mailbox/send`: Encrypts and routes a message to another teammate.
    *   `GET /mailbox/receive`: Retrieves and decrypts messages for the caller.
    *   `PATCH /tasks/{id}`: Updates task state with cryptographic signature validation.
*   **Data Storage/State:**
    *   Session-bound key-value store for public keys of active teammates.
    *   Encrypted SQLite backend for the Shared Task List.

## 5. Alternatives Considered
*   **Plaintext Shared State:** Rejected due to the risk of "PASI" (Protocol-Agnostic State Injection) where a low-trust subagent pollutes a high-trust reasoning loop.
*   **Framework-Specific Bridges:** Rejected because they don't solve the "Heterogeneous Swarm" problem (e.g., OpenClaw agents can't natively talk to Claude teammates).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** T2T implements the "Mailbox Integrity" strategic pivot. It ensures that even if a subagent is hijacked, it cannot coerce its teammates into unauthorized actions via message injection.
*   **Observability:** The `Inter-Agent Mailbox Monitor` provides a visual audit trail of all encrypted exchanges.

## 7. Evolutionary Changelog
*   **2026-05-22:** Initial Document Creation.

### Update: 2026-05-25 - Introducing Asynchronous Mailbox Sharding (AMS)
**Context:** Today's market sync revealed "Mailbox Lock" bottlenecks in horizontal swarms with 10+ teammates. The monolithic encrypted mailbox model is causing significant latency during peak coordination.
**Architecture Adjustment:**
*   Deprecating the single encrypted SQLite mailbox backend in Section 4.
*   Introducing **Asynchronous Mailbox Sharding (AMS)**. Every teammate-to-teammate pair now utilizes a dedicated, task-bound shard.
*   Implementing a lock-free queue for inter-shard synchronization.
**Security Impact:** Enhances isolation by ensuring a compromise of one shard doesn't expose the metadata or throughput of unrelated teammate coordination.
