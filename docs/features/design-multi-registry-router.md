# Design Doc: Multi-Registry Capability Router
**Status:** Draft
**Created:** 2026-04-08

## 1. Context and Scope
As the ecosystem matures, tool capabilities are no longer confined to a single local configuration. Gemini CLI v0.36 has introduced multi-registry support, where agents discover tools from local filesystems, organizational GitHub repositories, and private enterprise registries simultaneously. MCP Any needs a unified routing layer that can aggregate these disparate sources and provide a single, prioritized discovery bus to the agent, while managing naming collisions and trust boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate tool schemas from multiple independent local and remote registries.
    * Implement a "Trust-Weighted Resolution" policy to handle naming collisions across registries.
    * Provide a standardized "Registry Discovery" API for agents to query specific tiers (e.g., `local`, `org`, `community`).
    * Support dynamic registry mounting (e.g., "Mount this GitHub repo as a temporary tool registry").
* **Non-Goals:**
    * Syncing tool code between registries (we only route to the existing endpoint).
    * Enforcing a single global naming standard (we use namespaces to handle diversity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Full-Stack Agent Developer
* **Primary Goal:** Use a mix of local "Experimental" tools and enterprise-approved "Production" tools without manual config merging.
* **The Happy Path (Tasks):**
    1. The developer configures MCP Any with two registries: `local_fs` (high priority) and `org_registry` (base).
    2. The developer registers a custom `test_tool` in `local_fs`.
    3. The agent requests a list of all tools.
    4. The Multi-Registry Router aggregates tools from both sources.
    5. If a tool named `deploy` exists in both, the router returns the `local_fs:deploy` version due to higher priority.
    6. The developer mounts a partner's GitHub repo via the UI: `mcpany registry addPartner --url github.com/partner/tools`.
    7. The agent immediately "sees" the partner's tools in the discovery bus.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [Discovery Bus] -> [Multi-Registry Router] -> {Local Registry, Org Registry, Partner Registry}`
* **APIs / Interfaces:**
    * `Router.aggregate(query)`: Resolves a discovery request across all mounted registries.
    * `Router.mount_registry(type, config)`: Dynamically adds a new source of capabilities.
    * `Router.set_priority(registry_id, weight)`: Configures collision resolution policy.
* **Data Storage/State:**
    * Registry configurations and priority weights are stored in the core MCP Any config store. Discovery results are cached with a registry-specific TTL.

## 5. Alternatives Considered
* **Flat File Merging**: Rejected as it doesn't support dynamic/remote registries or real-time priority shifts.
* **Global UUIDs**: Rejected as it breaks existing natural-language tool discovery patterns used by LLMs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tools from "Community" registries are automatically quarantined in the "Discovery Sandbox" before exposure.
* **Observability:** The UI provides a "Registry Topology" view showing which tools originated from which source.

## 7. Evolutionary Changelog
* **2026-04-08:** Initial Document Creation.
