# Design Doc: JIT Permission Broker
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agent swarms become more autonomous (e.g., OpenClaw's self-healing loops), they frequently encounter "Permission Deadlocks" where a subagent requires a capability (e.g., `fs:write` to a specific directory) that was not explicitly granted by the parent agent. Currently, this requires human intervention, stalling the workflow. The JIT Permission Broker aims to solve this by providing a dynamic, automated adjudication process for on-demand privilege escalation.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable agents to request temporary, scoped permission elevations.
    * Use "High-Level Intent" from the Shared KV Store to adjudicate requests.
    * Provide a clear audit trail for all auto-escalated permissions.
    * Support "Time-to-Live" (TTL) for escalated tokens to ensure they are temporary.
* **Non-Goals:**
    * Eliminating the human-in-the-loop (HITL) for high-risk actions (e.g., deleting root filesystems).
    * General-purpose identity management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous DevOps Swarm Orchestrator.
* **Primary Goal:** Allow a "Fixer" agent to temporarily escalate permissions to write a patch to a repository after finding a bug.
* **The Happy Path (Tasks):**
    1. The Fixer agent attempts to write a file and receives a `403 Forbidden` from MCP Any.
    2. The Fixer agent calls the `request_elevation` tool, specifying the required capability (`fs:write:/repo/patch.py`) and the reason.
    3. The JIT Permission Broker retrieves the "Active Intent" from the session's Blackboard (Shared KV Store).
    4. The Broker verifies that "Fixing bug in repo" is the authorized high-level intent.
    5. The Broker issues a temporary, scoped token for `fs:write:/repo/patch.py`.
    6. The Fixer agent retries the write operation successfully.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: The Policy Firewall intercepts the `request_elevation` call.
    - **Adjudication**: The Broker queries the `Blackboard` for current session intent and compares it against the `Policy Engine` rules.
    - **Token Issuance**: If approved, a new `CapabilityToken` is generated with a short TTL (e.g., 5 minutes) and injected into the agent's session context.
* **APIs / Interfaces:**
    - `mcp.request_elevation(capability: string, reason: string)`: New system-level MCP tool.
* **Data Storage/State:** Escalation logs and temporary tokens are stored in the `Shared KV Store`.

## 5. Alternatives Considered
* **Static Over-Provisioning**: Giving agents more permissions than they need upfront. *Rejected* because it violates the Principle of Least Privilege.
* **Pure HITL**: Always asking a human. *Rejected* because it breaks autonomy in large-scale swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Escalation Hijacking" is a primary concern. Adjudication must use a separate, higher-tier LLM for reasoning than the one requesting escalation.
* **Observability:** Escalations are flagged in the UI with a distinct visual style in the Agent Chain Tracer.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
