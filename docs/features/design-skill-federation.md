# Design Doc: Federated Skill Subscription Adapter

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The "OpenClaw" ecosystem has popularized decentralized tool registries (Skills). Users want to dynamically "install" these tools into their agent sessions without manually editing configuration files or restarting the gateway. MCP Any needs a secure, audited way to fetch, validate, and proxy these federated skills.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Dynamic registration of MCP servers from remote skill registries (e.g., OpenClaw, SkillHub).
    *   Zero-Trust sandboxing for federated tools (automated capability restriction).
    *   Automated secret mapping for acquired skills.
    *   Audit logging for all "Skill Subscription" actions.
*   **Non-Goals:**
    *   Hosting a centralized skill registry (MCP Any is the client/gateway).
    *   Charging for skills (this is handled by the registry provider).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local AI Enthusiast using OpenClaw.
*   **Primary Goal:** Add a "Web Search" skill discovered in the community registry to their local MCP Any instance.
*   **The Happy Path (Tasks):**
    1.  User finds a skill URI (e.g., `skill://registry.openclaw.ai/web-search-v4`).
    2.  User tells their agent: "Install the web-search-v4 skill."
    3.  Agent calls `mcpany_subscribe_skill(uri="...")`.
    4.  MCP Any fetches the skill manifest, validates its cryptographic signature (Provenance Attestation), and checks permissions.
    5.  MCP Any prompts the user for required secrets (if not in Secret Mesh).
    6.  The skill is instantly available as a new toolset in the session.

## 4. Design & Architecture
*   **System Flow:**
    - **Subscription Handler**: Listens for `mcpany_subscribe_skill` calls.
    - **Manifest Fetcher**: Downloads and parses the skill's MCP configuration.
    - **Sandbox Engine**: Applies "Default-Deny" policies to the new toolset based on the skill's requested scopes.
    - **Dynamic Upstream**: A new type of Upstream that can be instantiated at runtime without a global config reload.
*   **APIs / Interfaces:**
    ```json
    {
      "method": "tools/call",
      "params": {
        "name": "mcpany_subscribe_skill",
        "arguments": {
          "uri": "string",
          "auto_approve_scopes": "boolean"
        }
      }
    }
    ```
*   **Data Storage/State:** Subscriptions are persisted in the `MCPANY_DB_PATH` in the `skill_subscriptions` table.

## 5. Alternatives Considered
*   **Static Manifest Files**: Forcing users to download YAML files and put them in a folder. *Rejected* because it breaks the "autonomous agent" workflow.
*   **Direct Plugin Execution**: Allowing skills to run as arbitrary binaries. *Rejected* due to extreme security risks; skills must be proxied through standard MCP transports (Stdio/HTTP) with strict isolation.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All federated skills are restricted to a "Minimal Capability Set" by default. Users must explicitly grant higher permissions (e.g., filesystem access).
*   **Observability:** Each skill has a dedicated health and performance dashboard.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
