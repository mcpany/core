# Design Doc: Autonomous Tool Refinement (ATR) Proxy
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Traditional MCP tool definitions are static. When an agent encounters a schema mismatch or ambiguous parameter requirements, the tool call fails, requiring manual developer intervention to update the MCP server. OpenClaw's ATR proposal suggests that agents can provide feedback to refine these schemas. MCP Any, as the universal gateway, is perfectly positioned to act as a "Learning Proxy" that intercepts these failures and facilitates autonomous refinement.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept tool execution errors related to schema validation or upstream API rejection.
    * Capture agent-suggested "refinements" (e.g., parameter renaming, type coercion, description updates).
    * Provide a Human-in-the-loop (HITL) workflow for approving refined schemas.
    * Persist refined schemas as "Evolutionary Overrides" for specific MCP tools.
* **Non-Goals:**
    * Automatically modifying the source code of upstream MCP servers.
    * Allowing unverified autonomous schema changes (HITL is mandatory for security).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Developer / System Operator
* **Primary Goal:** Resolve a tool call failure caused by a confusing schema without writing new code.
* **The Happy Path (Tasks):**
    1. Agent calls `jira_create_issue` but fails because `summary` was expected as `title`.
    2. Agent returns an ATR-compatible error response proposing the mapping: `title -> summary`.
    3. MCP Any's ATR Proxy captures this proposal and creates a "Refinement Request".
    4. Operator receives a notification in the MCP Any UI.
    5. Operator approves the refinement.
    6. Subsequent calls to `jira_create_issue` automatically map `title` to `summary` via the ATR middleware.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> ATR Middleware (Intercept) -> Upstream Tool -> [Error] -> ATR Middleware (Capture Proposal) -> [HITL Queue]`
* **APIs / Interfaces:**
    * `GET /atr/proposals`: List pending refinements.
    * `POST /atr/proposals/{id}/approve`: Apply the refinement to the runtime config.
* **Data Storage/State:**
    * Refinements are stored as "Config Overrides" in the local filesystem/database, merged at runtime with the base MCP configuration.

## 5. Alternatives Considered
* **Direct Server Patching**: Modifying the upstream MCP server code. *Rejected* because MCP Any aims to be a "Universal Adapter" that works with existing, often third-party, servers.
* **Hardcoded Mappings**: Manually adding mappings to `config.yaml`. *Rejected* as it's slow and doesn't leverage the agent's intelligence in identifying the fix.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Refinements could be used for "Prompt Injection" if an agent proposes a mapping that bypasses a policy. HITL approval is the primary mitigation.
* **Observability**: Refined tools will be marked with an "Evolved" badge in the UI, showing the history of schema changes.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
