# Design Doc: Protocol-Native HITL Signaling
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As AI agents move from experimental chat bots to autonomous systems, the need for human oversight (Human-in-the-Loop) is critical for high-consequence actions (e.g., deleting a database, making a financial transaction). Current HITL implementations are often "Middleware hacks" that suspend a request outside the core protocol, leading to timeouts and state synchronization issues. MCP Any needs to solve this by making HITL a first-class state within the Universal Adapter protocol.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize a `PENDING_APPROVAL` status in JSON-RPC and gRPC tool responses.
    * Provide a mechanism for agents to receive a "Call to Action" (CTA) with the approval request.
    * Support asynchronous resumption of tool calls after approval/denial.
* **Non-Goals:**
    * Implementing the actual UI for approval (this is handled by the MCP Any UI).
    * Defining the business logic for *when* to trigger HITL (this is handled by the Policy Engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Ensure all `terraform destroy` commands require explicit human approval via the MCP Any dashboard before execution.
* **The Happy Path (Tasks):**
    1. The agent identifies a need to run a high-risk tool call.
    2. The Policy Engine intercepts the call and marks it as `REQUIRES_APPROVAL`.
    3. MCP Any returns a `PENDING_APPROVAL` response to the agent with a unique `request_id`.
    4. The agent informs the user that approval is pending.
    5. The user reviews the request in the MCP Any UI and clicks "Approve."
    6. MCP Any resumes the tool call and returns the final result to the agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCP Any: Tool Call (id: 123)
        MCP Any->>Policy Engine: Validate(Call)
        Policy Engine-->>MCP Any: Status: Requires Approval
        MCP Any-->>Agent: JSON-RPC Error/Result (code: -32001, status: PENDING_APPROVAL, req_id: "abc")
        User->>MCP Any UI: Approve(req_id: "abc")
        MCP Any UI->>MCP Any: Execute(req_id: "abc")
        MCP Any->>Upstream: Real Tool Call
        Upstream-->>MCP Any: Result
        MCP Any-->>Agent: Notify/Poll Result (id: 123)
    ```
* **APIs / Interfaces:**
    * **JSON-RPC Response Extension:**
      ```json
      {
        "jsonrpc": "2.0",
        "id": 123,
        "result": {
          "status": "PENDING_APPROVAL",
          "approval_request_id": "auth_token_01",
          "message": "This action requires manual approval by an admin."
        }
      }
      ```
* **Data Storage/State:**
    * Use the `Shared KV Store` (SQLite) to track `approval_request_id` states, associated tool arguments, and original caller metadata.

## 5. Alternatives Considered
* **Long-polling/Blocking:** Rejected because it ties up server resources and leads to client timeouts during long human delays.
* **Middleware-only (Status Quo):** Rejected because it doesn't allow the LLM to understand *why* the tool call hasn't returned yet, leading to hallucinations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Approval requests must be cryptographically bound to the session and requires MFA for "High" risk tiers.
* **Observability:** All approval actions (Approve/Deny/Timeout) are recorded in the Audit Log.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
