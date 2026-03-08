# Design Doc: Intent-Bound Governance Middleware
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the rise of Autonomous Tool Synthesis (ATS) in frameworks like OpenClaw, agents are now capable of generating their own tools to bridge API gaps. This introduces a "capability escalation" risk where a synthesized tool might perform actions outside the agent's original mandate. Intent-Bound Governance shifts security from static "Role-Based" permissions to dynamic, "Mission-Bound" constraints.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically link every tool call to a signed "Mission Intent" string.
    * Enforce "Intent-Aware" boundaries where a tool call is rejected if it doesn't align with the high-level goal.
    * Provide a secure sandbox for executing synthesized (ATS) tools.
* **Non-Goals:**
    * Automatically generating the "Mission Intent" (this must be provided by the parent/orchestrator).
    * Validating the *semantic* truth of an intent (we trust the signer of the intent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Ensure a sub-agent synthesized to "read logs" cannot use its environment access to "delete files."
* **The Happy Path (Tasks):**
    1. Orchestrator starts a session with an "Intent Signature": `{"intent": "read-system-logs", "signer": "admin-key"}`.
    2. Sub-agent synthesizes a `log-reader` tool.
    3. Sub-agent calls `log-reader` via MCP Any.
    4. MCP Any verifies the tool call against the "read-system-logs" intent.
    5. MCP Any allows the read, but would block any `fs:write` attempt from the same synthesized tool.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [Intent Middleware] -> [ATS Sandbox] -> Tool Execution -> Policy Engine`
* **APIs / Interfaces:**
    * `X-MCP-Intent`: A new mandatory header for sessions under strict governance.
    * `X-MCP-Intent-Sig`: Cryptographic signature of the intent.
* **Data Storage/State:**
    * Intent metadata is stored in the session context and cached in the Policy Engine's "Intent Registry."

## 5. Alternatives Considered
* **Strict Whitelisting**: Rejected because it breaks Autonomous Tool Synthesis (ATS) which requires dynamic tool creation.
* **Manual HITL for every ATS tool**: Rejected as it's too slow for autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Intent-Bound" model ensures that even if an agent is compromised, its blast radius is limited to the current signed mission.
* **Observability:** Audit logs will explicitly link every tool execution to the specific "Mission Intent" that authorized it.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
