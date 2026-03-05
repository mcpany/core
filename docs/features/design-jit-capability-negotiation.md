# Design Doc: JIT Capability Negotiation Middleware
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
With the evolution of dynamic subagent frameworks like OpenClaw, static permission sets are becoming a bottleneck. Agents often don't know the full scope of required permissions at startup, leading to either "Permission Denied" errors or insecure "Over-Permissioning." This middleware enables agents to request specific capabilities just-in-time (JIT) during execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable agents to programmatically request additional tool permissions via a standardized MCP-like interface.
    * Provide a hook for Human-in-the-Loop (HITL) or Policy-Engine approval of escalation requests.
    * Maintain session-bound permission state without requiring a full service reload.
* **Non-Goals:**
    * Automating the approval process (the "Decision" remains with a human or a high-level security policy).
    * Persistent permission changes (JIT escalations are transient and session-bound).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Subagent
* **Primary Goal:** Access a restricted `filesystem:write` tool after discovering it's necessary for a task.
* **The Happy Path (Tasks):**
    1. Subagent attempts to call `write_file` and receives a `403 Forbidden` but with a `Capability-Request-ID`.
    2. Subagent calls the `negotiate_capability` tool with the Request ID and justification.
    3. MCP Any suspends the request and prompts the user/policy engine.
    4. Upon approval, the subagent's session token is dynamically updated with the new capability.
    5. Subagent retries the `write_file` call and succeeds.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Tool Call (Fails) -> Capability Negotiation Tool -> Policy Engine/User -> Session Token Update -> Retry`
* **APIs / Interfaces:**
    * `mcp_negotiate_capability(request_id: string, justification: string)`: The core tool exposed to agents.
    * `X-MCP-Capability-Request`: A new header for communicating negotiation state.
* **Data Storage/State:**
    * Pending and approved escalations are stored in the session's in-memory state.

## 5. Alternatives Considered
* **Static Profile Switching**: Too slow and disruptive for real-time agent swarms.
* **Wildcard Permissions**: Explicitly rejected due to Zero Trust principles.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Escalations must be tied to a parent's "Max Possible Scope" and require cryptographic attestation or human approval.
* **Observability:** All negotiation attempts, approvals, and denials are logged in the Audit Log.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
