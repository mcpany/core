# Design Doc: Swarm-to-Swarm (S2S) Handoff Protocol
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As AI agent deployments mature from single-agent tasks to multi-agent swarms (CrewAI, AutoGen), a new bottleneck has emerged: cross-cluster coordination. Currently, swarms are often siloed within a single network or framework. The S2S Handoff Protocol allows one swarm to securely hand over the entire state, history, and "responsibility" of a complex task to another swarm, potentially running on different infrastructure or managed by a different organization.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable secure, verifiable handoff of task ownership between agent swarms.
    * Standardize the "Context Package" format for swarm state (history, variables, tool results).
    * Provide a mechanism for "Asynchronous Handoff" where the receiving swarm might not be immediately available.
* **Non-Goals:**
    * Replacing existing A2A (Agent-to-Agent) protocols like those used within a single swarm.
    * Managing the internal orchestration of the receiving swarm.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Orchestrator
* **Primary Goal:** Hand off a "Market Research" task from a high-level "Strategy Swarm" to a specialized "Data Extraction Swarm" in a different VPC.
* **The Happy Path (Tasks):**
    1. Strategy Swarm identifies the need for deep data extraction.
    2. Strategy Swarm calls the `handoff_to_swarm` tool provided by MCP Any.
    3. MCP Any packages the current task context, signs it, and queues it for the Data Extraction Swarm.
    4. Data Extraction Swarm's gateway (MCP Any) receives the package, verifies the signature, and initializes the local swarm with the inherited context.
    5. The Data Extraction Swarm completes the task and "hands back" the result via the same protocol.

## 4. Design & Architecture
* **System Flow:**
    `[Swarm A] -> [MCP Any Gateway A] -> (Encrypted Context Package) -> [MCP Any Gateway B] -> [Swarm B]`
* **APIs / Interfaces:**
    * `POST /v1/s2s/handoff`: Initiates a handoff.
    * `GET /v1/s2s/mailbox`: Retrieves pending handoff packages for a swarm.
* **Data Storage/State:** Handoff packages are stored in a persistent "Stateful Buffer" within MCP Any until successfully acknowledged by the receiving gateway.

## 5. Alternatives Considered
* **Direct A2A calls:** Rejected because it requires synchronous connectivity and doesn't handle the "Swarm-level" state inheritance well.
* **Shared Database:** Rejected due to security concerns across different administrative domains; S2S requires explicit, signed transfers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All handoff packages are cryptographically signed by the source gateway. The destination gateway must have the source's public key in its "Trusted Swarms" list.
* **Observability:** MCP Any logs "Handoff Initiated," "Package Delivered," and "Acknowledgment Received" events with full trace IDs.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
