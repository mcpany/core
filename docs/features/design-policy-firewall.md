# Design Doc: Policy Firewall
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents evolve from simple chatbots to autonomous systems capable of executing complex workflows, the security risks have scaled exponentially. Recent vulnerabilities like LangGrinch (CVE-2025-68664) and "Clinejection" demonstrate that simple tool allowlists are insufficient. Agents need a robust, granular, and intent-aware security perimeter.

The **Policy Firewall** is a vendor-agnostic middleware for MCP Any that enforces fine-grained access control policies on every tool call and response. It acts as a "Zero Trust" gateway between the LLM/Agent and the underlying capabilities.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified policy enforcement point (PEP) for all MCP tool interactions.
    * Support industry-standard policy languages (Rego/OPA, CEL).
    * Enable granular, resource-level permissions (e.g., `fs:read:/tmp` vs `fs:read:/etc`).
    * Implement "Intent-Aware" verification where policies are evaluated against the parent agent's high-level goal.
    * Support synchronous suspension for Human-in-the-Loop (HITL) approvals.
* **Non-Goals:**
    * Replacing existing OS-level security (e.g., AppArmor, SELinux).
    * Storing user identities (authentication); it focuses on capability-based authorization.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer / Enterprise Admin.
* **Primary Goal:** Prevent an autonomous coding agent from reading sensitive environment secrets while allowing it to read project source files.
* **The Happy Path (Tasks):**
    1. Admin defines a Rego policy that denies any `filesystem` tool call where the path contains `.env` or `config/secrets`.
    2. Admin registers the policy with the MCP Any Policy Firewall.
    3. An AI agent (e.g., Claude Code) attempts to read `.env` via an MCP tool.
    4. Policy Firewall intercepts the request, evaluates the Rego policy, and blocks the call.
    5. Policy Firewall returns a "Security Violation" error to the agent and logs the event for auditing.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCP Any: Tool Call (Request)
        MCP Any->>Policy Firewall: Intercept Request
        Policy Firewall->>OPA Engine: Evaluate Policy (Rego)
        OPA Engine-->>Policy Firewall: Allow/Deny/Suspend
        alt Allow
            Policy Firewall->>MCP Tool: Execute
            MCP Tool-->>Policy Firewall: Response
            Policy Firewall->>Hardened Serialization: Sanitize Response
            Hardened Serialization-->>Agent: Sanitized Response
        alt Deny
            Policy Firewall-->>Agent: Security Violation Error
        alt Suspend
            Policy Firewall->>HITL Queue: Request Approval
            HITL Queue-->>Policy Firewall: User Decision
            Note over Policy Firewall: Resume or Deny
        end
    ```
* **APIs / Interfaces:**
    * `POST /v1/policies`: Register/Update a policy.
    * `GET /v1/audit/logs`: Retrieve security audit events.
    * Middleware Hook: `OnBeforeToolCall(context, request) -> (response, error)`
* **Data Storage/State:**
    * Policies are stored in an internal SQLite database or a version-controlled Git repository.
    * Audit logs are streamed to the observability layer.

## 5. Alternatives Considered
* **Native Agent Policies (Gemini/Claude)**: Rejected as a primary solution because they are vendor-locked and inconsistent across different agent frameworks. MCP Any's goal is to be the *universal* bridge.
* **Hardcoded Allow/Deny Lists**: Rejected as too brittle for complex, context-dependent security requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The firewall itself must be isolated. We use "Strict Seatbelt Profiles" to ensure the OPA engine cannot be bypassed.
* **Observability**: Every policy evaluation is logged with full context (Agent ID, Intent, Tool, Arguments, Decision).

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
