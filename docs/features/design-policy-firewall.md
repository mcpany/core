# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
As AI agents become more autonomous, they increasingly require access to sensitive infrastructure (shell, databases, file systems). Current safety measures are often hardcoded or nonexistent in frameworks like OpenClaw. MCP Any needs a flexible, declarative way to intercept and validate tool calls based on context, identity, and intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a Zero Trust execution environment for MCP tools.
    * Support Rego (OPA) and CEL (Common Expression Language) for policy definitions.
    * Allow dynamic policy updates without server restarts.
    * Log all policy decisions (Allow/Deny) for auditability.
* **Non-Goals:**
    * Replacing existing authentication (AuthN) - this is an Authorization (AuthZ) layer.
    * Real-time monitoring of subprocess resource usage (CPU/RAM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Administrator
* **Primary Goal:** Prevent an autonomous agent from executing destructive commands while allowing safe operations in a specific directory.
* **The Happy Path (Tasks):**
    1. Admin defines a policy: `allow_exec = true if input.args.path.startswith("/tmp/sandbox")`.
    2. Agent attempts `call_tool(exec, {path: "/tmp/sandbox/script.sh"})`.
    3. Policy Firewall intercepts, evaluates against input.
    4. Policy Firewall returns `ALLOW`.
    5. Tool executes.

## 4. Design & Architecture
* **System Flow:**
    `Client Request` -> `Auth Middleware` -> **`Policy Firewall (Intercept)`** -> `Adapter Execution` -> `Response Transformation`.
* **APIs / Interfaces:**
    * New configuration block: `policy_engine: { type: "rego", path: "./policies/*.rego" }`.
* **Data Storage/State:**
    Policies are loaded from disk or a remote configuration store. Evaluation is stateless per request.

## 5. Alternatives Considered
* **Hardcoded Filters:** Rejected due to lack of flexibility for enterprise users.
* **External OPA Sidecar:** Rejected for performance reasons and to minimize "binary fatigue".

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Policies default to `DENY` if no rules match.
* **Observability:** Every decision is logged to the `audit_log` with the matching rule ID.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation.
