# Design Doc: Subagent Tool Filtering Middleware (STFM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of autonomous specialist agents, the risk of "Capability Creep" has become a significant security threat. A subagent designed for "Log Analysis" may inadvertently discover and execute "Database Drop" tools if the tool registry is not strictly filtered. Gemini CLI recently introduced subagent tool filtering to address this.

MCP Any must implement Subagent Tool Filtering Middleware (STFM) as a core security layer. STFM will perform real-time, semantic analysis of an agent's capability card and mission role, pruning any tools that do not align with the verified "Mission Root" intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce granular, role-based tool filtering for all subagent spawns.
    * Neutralize "Privilege Escalation" via autonomous tool discovery.
    * Provide a hardware-attested manifest of "Active Tools" per subagent session.
* **Non-Goals:**
    * Replacing the main Policy Firewall (STFM is a pre-discovery and pre-invocation filter).
    * Modifying tool schemas (it only controls visibility and access).

## 3. Critical User Journey (CUJ)
* **User Persona:** Zero-Trust Swarm Architect
* **Primary Goal:** Ensure that a "Frontend Auditor" subagent can only see and call read-only UI tools, even if the parent has full system access.
* **The Happy Path (Tasks):**
    1. Parent agent spawns a subagent with the role "Frontend Auditor."
    2. MCP Any intercepts the spawn request and applies the STFM filtering policy.
    3. STFM queries the Mission Root manifest and prunes all non-UI tools from the subagent's view.
    4. Subagent attempts to discover tools and only receives capability cards for authorized UI tools.
    5. Subagent attempts to call a forbidden "Shell" tool; STFM blocks the call at the gateway layer based on the filtered manifest.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Parent[Parent Agent] -->|Spawn Subagent: Frontend Auditor| Gateway[MCP Any Gateway]
        Gateway -->|Verify Role| STFM[Subagent Tool Filtering Middleware]
        STFM -->|Query Manifest| HAMM[Hardware-Attested Mission Manifest]
        STFM -->|Filter Capabilities| Subagent[Subagent: Frontend Auditor]
        Subagent -->|Restricted Tool View| Registry[Tool Registry]
    ```
* **APIs / Interfaces:**
    * `POST /v1/subagent/filter`: Apply a filtering policy to a specific session.
    * `GET /v1/subagent/capabilities`: Retrieve the filtered manifest for an active subagent.
* **Data Storage/State:**
    * Filtered manifests are stored as session-bound, immutable shards in the DME Broker.

## 5. Alternatives Considered
* **Static RBAC**: Rejected because agent missions are dynamic and require "Just-in-Time" filtering based on reasoning context.
* **Post-Call Blocking Only**: Rejected because exposing sensitive tool schemas to a subagent can lead to instruction injection and reconnaissance attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Filtering is enforced at the kernel/gateway layer, independent of the subagent's internal state.
* **Observability:** Filtered events (pruned tools) are logged to the Stylometric Mesh Dashboard to detect intent drift.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
