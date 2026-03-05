# Design Doc: Verified Skill Proxy (VSP)
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The rapid growth of the MCP ecosystem has led to the emergence of "Skill Marketplaces" like ClawHub. While these enable rapid feature expansion for agents, they also introduce significant supply chain risks. Malicious tool definitions (skills) can be used to exfiltrate sensitive data, environment variables, or perform unauthorized actions. MCP Any, as a universal gateway, is uniquely positioned to act as a "Sanitizing Proxy" between agents and potentially untrusted MCP servers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept and scan MCP tool definitions (schemas) against a database of known malicious patterns.
    *   Implement "Signature Verification" for tools originating from trusted registries.
    *   Provide a "Sandboxed Execution" hint to LLMs for tools that fail high-confidence verification.
    *   Enable users to "Quarantine" suspicious tools before they are exposed to the agent.
*   **Non-Goals:**
    *   Building a complete Antivirus for LLMs (focus is on the tool definition and protocol-level behavior).
    *   Replacing the Policy Engine (VSP is a pre-processor for the Policy Engine).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer installing a new "Skill" from a community marketplace.
*   **Primary Goal:** Ensure the new tool doesn't contain hidden data exfiltration logic.
*   **The Happy Path (Tasks):**
    1.  Developer adds a new MCP server URL to `config.yaml`.
    2.  MCP Any connects to the server and retrieves tool definitions.
    3.  The **Verified Skill Proxy** intercepts the response and checks the tool schemas against a real-time threat feed.
    4.  VSP detects a suspicious `post_install_script` or an unusual `exfiltrate_env` parameter.
    5.  MCP Any blocks the tool and notifies the user via the UI/CLI.
    6.  User reviews the "Quarantine" report and chooses to either delete the service or override the block.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Hook**: VSP registers a hook in the `ServiceRegistry` that fires whenever new tools are discovered.
    - **Pattern Matcher**: A regex and AST-based scanner that looks for "Code Injection" or "Data Leakage" patterns in tool descriptions and input schemas.
    - **Reputation Service**: A client that syncs with an external "MCP Threat Feed" (e.g., maintained by the MCP Any community).
*   **APIs / Interfaces:**
    - `VerifiedSkillProxyInterface`: Methods for `ScanTool(definition)` and `VerifySignature(definition, signature)`.
    - New Admin API: `GET /api/v1/quarantine` - Returns blocked tools and reasons.
*   **Data Storage/State:** A local cache of known safe/malicious tool hashes in the SQLite database.

## 5. Alternatives Considered
*   **Manual Review Only**: Forcing users to manually approve every new tool. *Rejected* as it doesn't scale for complex agent swarms.
*   **LLM-based Scanning**: Using an LLM to "audit" the tool code. *Rejected* as too slow and expensive for a real-time proxy layer, though it could be an "Advanced" option.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** VSP is a critical component of supply chain security. It ensures that the agent's "Trust Boundary" extends to the tools it consumes.
*   **Observability:** The UI should show a "Security Audit" badge on every tool, indicating its VSP status (Verified, Unverified, Suspicious).

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
