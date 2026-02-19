# Design Doc: Zero Trust Policy Firewall
**Status:** Draft
**Created:** 2026-02-19

## 1. Context and Scope
Autonomous AI agents are increasingly being granted broad access to sensitive systems (APIs, databases, filesystems). However, existing MCP implementations lack a granular security layer, leading to vulnerabilities like RCE and command injection (e.g., CVE-2026-25253 in OpenClaw).

MCP Any needs a "Policy Firewall" that can intercept, evaluate, and potentially block or redact tool calls based on declarative security policies. This ensures that agents operate within strictly defined boundaries, even if the underlying agent logic is compromised.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a declarative policy engine using Rego (Open Policy Agent) or CEL (Common Expression Language).
    * Evaluate tool calls in real-time based on agent identity, tool name, and payload.
    * Support "Dry Run" and "Audit Only" modes for policy testing.
    * Provide a mechanism for "Soft Blocks" that trigger Human-in-the-Loop (HITL) approval.
* **Non-Goals:**
    * Replacing existing authentication (API Keys, OAuth) - the firewall acts as a secondary authorization layer.
    * Implementing the LLM-side logic for choosing tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an AI agent from executing `DELETE` operations on production databases without explicit human approval, while allowing `GET` operations.
* **The Happy Path (Tasks):**
    1. Architect defines a Rego policy file (`policy.rego`) that denies any tool call where `method == "DELETE"` and `env == "prod"`.
    2. Architect configures MCP Any to use this policy file in `config.yaml`.
    3. An AI agent attempts to call a `delete_user` tool.
    4. MCP Any intercepts the call, evaluates it against the Rego policy.
    5. The policy returns a "DENY" with a requirement for "HITL_APPROVAL".
    6. MCP Any suspends the call and notifies the architect via the UI/Webhook.
    7. Architect approves the call; MCP Any executes it and returns the result to the agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|Tool Call| Server[MCP Any Server]
        Server -->|Payload| PE[Policy Engine (Rego/CEL)]
        PE -->|Decision: ALLOW/DENY/HITL| Server
        Server -->|If DENY| Error[Error Response]
        Server -->|If HITL| HITL[HITL Middleware]
        Server -->|If ALLOW| Adapter[Upstream Adapter]
        Adapter -->|Execute| Upstream[Upstream Service]
        Upstream -->|Result| Server
        Server -->|Result| Agent
    ```
* **APIs / Interfaces:**
    * New configuration block in `config.yaml`:
      ```yaml
      policyFirewall:
        engine: "rego"
        policyPath: "./policies/main.rego"
        defaultAction: "deny"
      ```
    * Middleware interface: `Evaluate(ctx context.Context, req *ToolRequest) (*PolicyDecision, error)`

* **Data Storage/State:**
    * Policies are stored as `.rego` or `.cel` files.
    * Evaluation results and audit logs are stored in the existing SQLite database.

## 5. Alternatives Considered
* **Hardcoded Rules:** Rejected because they are not flexible enough for diverse enterprise needs.
* **LLM-based Moderation:** Rejected because it is non-deterministic and can be bypassed by prompt injection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The policy engine itself must be isolated to prevent "policy injection". Policies should be signed and verified.
* **Observability:** Every policy decision (Allow/Deny/HITL) must be logged in the Audit Trail with the associated trace ID.

## 7. Evolutionary Changelog
* **2026-02-19:** Initial Document Creation.
