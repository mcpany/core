# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2025-02-17

## 1. Context and Scope
As AI agents become more autonomous, the risk of "prompt injection" or "rogue subagent" actions increases. MCP Any needs a robust way to intercept and validate every tool call against a predefined security policy before it reaches the upstream service.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a Rego or CEL based engine for defining tool-call policies.
    * Support granular access control based on tool name, arguments, and agent identity.
    * Enable real-time blocking of unauthorized or dangerous tool calls.
* **Non-Goals:**
    * Replacing the authentication layer (this is an authorization layer).
    * Providing a full-blown IAM system.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Engineer / Agent Developer.
* **Primary Goal:** Prevent an agent from calling `delete_database` unless it has a specific "admin" scope.
* **The Happy Path (Tasks):**
    1. Define a Rego policy file (`policy.rego`).
    2. Register the policy in the `mcpany` configuration.
    3. Agent attempts a `delete_database` call.
    4. Policy Firewall intercepts the call, evaluates it against the Rego engine, and allows it only if the scope matches.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `MCP Any Server` -> `Policy Firewall Middleware` -> `Rego Engine (Evaluate)` -> `Upstream Adapter` -> `Upstream Service`
* **APIs / Interfaces:**
    * `PolicyEngine` interface in Go with an `Evaluate(context, request) Result` method.
* **Data Storage/State:**
    * Policies are loaded from disk as `.rego` or `.yaml` files.

## 5. Alternatives Considered
* **Hardcoded Policies**: Rejected because it's not flexible enough for diverse user needs.
* **External OPA Server**: Rejected to minimize latency and operational complexity for local-first deployments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Policies must be read-only and immutable at runtime.
* **Observability:** Every policy evaluation (Allow/Deny) must be logged in the audit trail.

## 7. Evolutionary Changelog
* **2025-02-17:** Initial Document Creation.
