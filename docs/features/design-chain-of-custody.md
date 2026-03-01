# Design Doc: Chain of Custody (CoC) Middleware
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As multi-agent swarms (e.g., OpenClaw, CrewAI) become more complex, the delegation of tasks from a "lead" agent to "subagents" creates a security vacuum. Current policy engines typically only verify the identity of the immediate caller. The "Shadow Agent Chain" exploit demonstrated that a rogue or hallucinating subagent can spawn its own tool-calling loops that exceed the authority intended by the parent agent. MCP Any must provide a mechanism to cryptographically link every action back to a verified chain of intent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a cryptographic "Handover" protocol for agent-to-agent delegation.
    *   Ensure every tool call carries a "Chain of Custody" (CoC) token that represents the lineage of authority.
    *   Integrate CoC verification into the Policy Firewall to enforce "Intent-Scoped" permissions.
*   **Non-Goals:**
    *   Determining the "intent" itself (this is handled by the model/parent agent).
    *   Replacing traditional RBAC/ABAC (CoC is an additional layer of verification).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Architect.
*   **Primary Goal:** Prevent a "Refinement Subagent" from accessing sensitive environment variables that only the "Orchestrator" should see.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator agent initiates a subagent task via the A2A Bridge.
    2.  MCP Any's CoC Middleware generates a signed `CoC Token` that includes the parent's identity and the specific sub-task scope.
    3.  The subagent receives the token and includes it in all its tool calls.
    4.  The Policy Firewall interceptor validates the token's signature and ensures the requested tool is within the sub-task's scope.
    5.  The tool call is executed, and the CoC lineage is recorded in the audit log.

## 4. Design & Architecture
*   **System Flow:**
    - **Token Generation**: When an agent delegates via MCP Any, the middleware wraps the request and signs it using the instance's private key (attesting to the handover).
    - **Propagation**: CoC tokens are passed via standard MCP/A2A headers (`X-MCP-CoC-Chain`).
    - **Verification**: The Policy Firewall parses the chain, verifying each "hop" in the delegation.
*   **APIs / Interfaces:**
    - `coc/sign-handover`: Internal API for generating tokens.
    - `coc/verify-chain`: Middleware hook for tool call interception.
*   **Data Storage/State:** CoC tokens are short-lived and stateless (JWT-like), but the signatures are verified against the instance's identity stored in `~/.mcpany/id_ed25519`.

## 5. Alternatives Considered
*   **Centralized Session State**: Storing all delegation in a database. *Rejected* because it creates a bottleneck and is harder to scale in a federated mesh.
*   **LLM-Based Intent Verification**: Asking a "Judge" model if the call is okay. *Rejected* due to high latency and cost, though it could be a fallback.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** CoC is the "Proof of Intent" in the Zero Trust model. It prevents "Privilege Escalation via Hallucination."
*   **Observability:** The UI must visualize the CoC chain (e.g., "Orchestrator -> Researcher -> Coder") for every tool call in the trace viewer.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
