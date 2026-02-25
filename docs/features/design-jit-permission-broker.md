# Design Doc: JIT Permission Broker
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agent swarms (OpenClaw, CrewAI, etc.) become more autonomous, they frequently encounter "Permission Deadlocks." These occur when a subagent requires a capability (e.g., `fs:write`) that was not explicitly granted to its parent. Currently, this leads to task failure. The JIT Permission Broker aims to solve this by providing a mechanism for agents to request ephemeral, scoped permissions during execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized MCP-like interface for agents to request permission elevation.
    * Support auto-approval based on Rego/CEL policy rules.
    * Integrate with the HITL Middleware for human approval of sensitive requests.
    * Ensure all JIT permissions are ephemeral and automatically expire.
* **Non-Goals:**
    * Replacing the base Zero-Trust permission model (it extends it).
    * Providing a persistent permission management UI (it's for runtime escalation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Agent Swarm Orchestrator
* **Primary Goal:** Successfully complete a multi-step task that requires unexpected tool access without crashing.
* **The Happy Path (Tasks):**
    1. A subagent attempts to call `git:commit` but lacks the `repo:write` capability.
    2. The Policy Firewall blocks the call but returns a `PermissionDenied` error with a `request_jit_escalation` hint.
    3. The Orchestrator calls the `jit_request_permission` tool provided by MCP Any.
    4. The JIT Broker evaluates the request against Rego policies.
    5. The request is auto-approved because it aligns with the high-level task intent.
    6. The subagent retries the `git:commit` call and succeeds.
    7. After the session ends, the ephemeral capability is revoked.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [Policy Firewall (DENY)] -> [Orchestrator] -> [JIT Broker] -> [Policy Engine/HITL] -> [Policy Firewall (ALLOW)]`
* **APIs / Interfaces:**
    * `jit_request_permission(capability: string, duration: string, justification: string)`: Standardized tool exposed to agents.
    * `jit_approve(request_id: string)` / `jit_deny(request_id: string)`: Internal API for HITL/Policy Engine.
* **Data Storage/State:**
    * Ephemeral permissions are stored in the `Shared KV Store` (SQLite) with an expiration timestamp.

## 5. Alternatives Considered
* **Static Over-Provisioning**: Giving agents more permissions than they need. Rejected due to security risks (Violates Principle of Least Privilege).
* **Manual Configuration Update**: Requiring a human to update `config.yaml` and restart. Rejected due to friction and loss of autonomy.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Justification" field is logged but not trusted; only Rego policies and HITL define the security boundary. Prevents "Social Engineering" of the broker by the agent.
* **Observability:** All JIT requests, approvals, and denials are recorded in the audit logs with full context (agent ID, parent task, justification).

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
