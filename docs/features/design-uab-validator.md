# Design Doc: UAB 1.2 Authenticated Task Card Validator
**Status:** Draft
**Created:** 2026-03-19

## 1. Context and Scope
With the rise of the Universal Agent Bus (UAB) as a standard for inter-agent communication, there is a pressing need for a secure way to delegate tasks between agents that may not share a direct trust relationship. UAB 1.2 introduces "Authenticated Task Cards"—cryptographically signed objects that define a task, its scope, and the permissions granted to the performer. MCP Any must act as the "Certificate Authority" and "Validator" for these cards to enable secure cross-framework swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Verify the cryptographic signatures of UAB 1.2 Task Cards.
    * Enforce "Capability-Based" permissions based on the task card's scope.
    * Support task handoffs between disparate frameworks (e.g., OpenClaw to AutoGen).
    * Maintain a "Lineage Audit Trail" for every delegated task.
* **Non-Goals:**
    * Serving as a task marketplace (MCP Any is the infrastructure, not the broker).
    * Executing the tasks itself (MCP Any secures the communication).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Developer
* **Primary Goal:** Securely delegate a "Code Review" task from an OpenClaw agent to an AutoGen specialist without giving the specialist full access to the repo.
* **The Happy Path (Tasks):**
    1. Parent Agent (OpenClaw) generates a UAB 1.2 Task Card.
    2. Parent signs the card using its MCP Any-issued identity.
    3. Parent sends the Task Card to the Specialist (AutoGen) via MCP Any.
    4. MCP Any validates the signature and the scope (e.g., "Read-Only access to /src").
    5. Specialist attempts to call `write_file`.
    6. MCP Any blocks the call because it exceeds the Task Card's scope.

## 4. Design & Architecture
* **System Flow:**
    `Task Delegation` -> `Signature Verification` -> `Scope Attestation` -> `Policy Enforcement`
* **APIs / Interfaces:**
    * `/uab/v1.2/validate`: Endpoint for verifying task cards.
    * `X-UAB-Task-Card`: Header for propagating task context in tool calls.
* **Data Storage/State:**
    * Identity Store: Mapping agent framework IDs to public keys.
    * Active Task Registry: Tracking live task cards and their current depth.

## 5. Alternatives Considered
* **OIDC-based Delegation**: Rejected as too heavy for high-frequency agent-to-agent tool calls.
* **Simple API Keys**: Rejected because they don't support granular task-level scoping or lineage tracking.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Ensures that subagents cannot escalate their privileges beyond what was explicitly delegated.
* **Observability:** Detailed audit logs showing which agent delegated which task to whom.

## 7. Evolutionary Changelog
* **2026-03-19:** Initial Document Creation.
