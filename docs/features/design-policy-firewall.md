# Design Doc: Policy Firewall

**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
As AI agents gain the ability to execute complex tool chains and manage project-local configurations, the risk of unauthorized actions (e.g., recursive file deletion, credential exfiltration) increases. Current agent frameworks often rely on "all-or-nothing" permissions. The Policy Firewall provides a granular, capability-based security layer that intercepts all tool calls and configuration hooks, validating them against declarative policies before execution.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept and validate every tool call made by an agent.
    *   Support declarative policy languages (Rego/CEL) for complex rules.
    *   Enable "Intent-Bound Isolation" where permissions are restricted to a specific task scope.
    *   Sanitize and validate project-local configurations (e.g., `.claude/settings.json`) to prevent RCE.
    *   Provide a standard "Local-Only" default policy that requires explicit attestation for remote access.
*   **Non-Goals:**
    *   Replacing the underlying tool execution logic.
    *   Implementing a general-purpose IAM system for humans (this is for agents).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer using OpenClaw.
*   **Primary Goal:** Prevent an open-source agent from accessing `~/.ssh` or executing unauthorized `rm -rf` commands.
*   **The Happy Path (Tasks):**
    1.  Developer starts MCP Any with the Policy Firewall enabled.
    2.  Developer defines a policy: `allow fs:read if path.startsWith("/project/src")`.
    3.  OpenClaw attempts to read `~/.ssh/id_rsa`.
    4.  Policy Firewall intercepts the call, evaluates the Rego policy, and blocks the request.
    5.  MCP Any returns a `PermissionDenied` error to the agent and logs the attempt in the Security Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: Middleware layer in the tool execution pipeline.
    - **Evaluation Engine**: A stateless engine that executes Rego (OPA) or CEL rules.
    - **Context Enrichment**: The firewall injects caller metadata (Agent ID, Intent Token, Session ID) into the evaluation context.
*   **APIs / Interfaces:**
    - `ValidateToolCall(context, tool, args)`: Internal interface for the execution pipeline.
    - `SetPolicy(policy_id, rego_content)`: Administrative API for updating rules.
*   **Data Storage/State:** Policies are stored as versioned files or in a local database. Evaluation is stateless for performance.

## 5. Alternatives Considered
*   **Hardcoded Rules**: simple if/else blocks in the code. *Rejected* because it lacks the flexibility required for diverse agent ecosystems.
*   **OS-Level Sandboxing (Docker)**: running every tool in a container. *Rejected* as the primary mechanism due to performance overhead and complexity in sharing state, though it may be used as a secondary "Detached Sandbox" for automated hooks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The firewall itself must be tamper-proof. Policies should be signed.
*   **Observability:** Every block/allow decision must be logged with high fidelity for auditing and debugging.
*   **Performance:** Policy evaluation must be sub-millisecond to avoid impacting agent reasoning loops.

## 7. Evolutionary Changelog
*   **2026-03-11:** Initial Document Creation.
