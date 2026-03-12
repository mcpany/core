# Design Doc: Policy Firewall
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
As the Universal Agent Bus, MCP Any mediates all communications between agents and tools. As the number of tools and the complexity of agent behaviors grow, hardcoded permissions are insufficient. The Policy Firewall provides a declarative, granular, and dynamic way to control tool execution using Rego (Open Policy Agent) or CEL (Common Expression Language). It ensures "Zero Trust" by verifying every tool call against organizational and project-level security policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept every tool call and evaluate it against a set of policies.
    * Support granular rules based on agent identity, tool name, arguments, and environment context.
    * Provide "Safe-by-Default" policies that block dangerous operations (e.g., recursive filesystem deletion).
    * Enable "Project-Local" policy overrides that must be attested by the user.
* **Non-Goals:**
    * Implementing a full Identity Provider (IdP) – uses existing auth headers.
    * Modifying the underlying tool code – acts only as a proxy.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer at a Fintech company.
* **Primary Goal:** Prevent an agent from calling a "Database Query" tool on a production table unless specifically authorized for a session.
* **The Happy Path (Tasks):**
    1. Security Engineer defines a Rego policy in `policies/db-safety.rego`.
    2. Agent attempts to call `sql_query(query="DROP TABLE users")`.
    3. MCP Any intercepts the call and passes the tool name and arguments to the Policy Engine.
    4. The Policy Engine evaluates the Rego rule and returns `allow: false, reason: "Destructive SQL prohibited"`.
    5. MCP Any returns an error to the agent and logs the attempt in the Audit Log.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Gateway` -> `Middleware: Policy Engine` -> `Upstream Tool`
    1. **Context Enrichment**: The Policy Engine gathers metadata (Agent Identity, Intent Scope, Source Locality).
    2. **Rule Evaluation**: Compiles and runs Rego/CEL against the tool call payload + metadata.
    3. **Action Execution**:
        * `Allow`: Call proceeds.
        * `Deny`: Call is blocked, error returned.
        * `Suspend`: Trigger HITL Middleware for user approval.
* **APIs / Interfaces:**
    * `GET /v1/policies`: List active policies.
    * `POST /v1/policies/test`: Dry-run a tool call against the policy set.
* **Data Storage/State:**
    * Policies stored as `.rego` or `.yaml` (CEL) files in the `config/policies/` directory.

## 5. Alternatives Considered
* **Hardcoded Permission Groups**: Rejected; too rigid and doesn't scale to heterogeneous agent swarms.
* **Agent-Side Enforcement**: Rejected; agents can be bypassed or hallucinate their way around local rules.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The firewall is the core enforcer of the Zero Trust pillar. It must be "Fail-Closed" – if the engine fails, all tool calls are denied.
* **Observability**: Every policy decision is recorded with a unique `PolicyDecisionID` linked to the `TraceID`.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
