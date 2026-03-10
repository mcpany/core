# Design Doc: Swarm Attestation Headers
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As AI agent ecosystems move towards decentralized swarms (e.g., OpenClaw Swarm v2), agents are increasingly discovering and invoking tools directly from each other. This "Shadow MCP" pattern bypasses central security gateways, making it impossible to verify if a tool call was actually authorized by the root orchestrator. MCP Any needs to provide a way to cryptographically verify the "Chain of Custody" for every tool call within a swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a standardized header format for tool call attestation.
    * Allow root orchestrators to sign "Intent Tokens" that subagents must present when calling tools.
    * Enable MCP Any to verify the entire signature chain back to a trusted root.
    * Support rotating keys and multiple trusted roots.
* **Non-Goals:**
    * Implement the agents' internal reasoning for when to call a tool.
    * Replace existing transport-level encryption (TLS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw)
* **Primary Goal:** Ensure that a "Write File" tool call from a sub-sub-agent was explicitly authorized by the main orchestrator for a specific task.
* **The Happy Path (Tasks):**
    1. The Orchestrator initiates a task and generates a signed `RootIntentToken`.
    2. The Orchestrator passes this token to Subagent A.
    3. Subagent A needs to call a tool via Subagent B. It wraps the `RootIntentToken` and its own signature into a `SwarmAttestationHeader`.
    4. Subagent B receives the call and forwards it to MCP Any.
    5. MCP Any verifies the `SwarmAttestationHeader`, ensuring the chain leads back to the Orchestrator's trusted key and the intent matches the requested tool.
    6. MCP Any executes the tool and returns the result.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant O as Orchestrator
        participant A as Subagent A
        participant B as Subagent B
        participant M as MCP Any
        O->>O: Sign Intent(task_id, allowed_tools) -> Token_O
        O->>A: Task + Token_O
        A->>A: Sign(Token_O, call_details) -> Token_A
        A->>B: Call(tool, Token_A)
        B->>M: CallTool(tool, headers: {X-MCP-Attestation: Token_A})
        M->>M: Verify Chain(Token_A -> Token_O -> RootKey)
        M->>M: Validate Intent(tool, allowed_tools)
        M->>B: ToolResult
        B->>A: Result
    ```
* **APIs / Interfaces:**
    * New middleware `AttestationValidator` in the request pipeline.
    * Support for `X-MCP-Attestation` header in JSON-RPC metadata.
* **Data Storage/State:**
    * Trusted Public Keys stored in MCP Any's Service Registry / Policy Engine.

## 5. Alternatives Considered
* **Centralized Authorization Service:** Rejected because it introduces a single point of failure and high latency for decentralized swarms.
* **Mutual TLS (mTLS) between all agents:** Rejected due to the operational complexity of managing certificates for thousands of ephemeral subagents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevents unauthorized tool execution by rogue subagents or compromised swarm members.
* **Observability:** Attestation chains will be logged in audit trails, providing full provenance for every action.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
