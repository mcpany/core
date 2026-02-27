# Design Doc: Enterprise IAM Bridge (OIDC/SAML to MCP)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As MCP Any moves into enterprise environments (e.g., following the release of Claude Cowork Connectors), tool access cannot remain as a single "developer-local" permission. We need to bridge existing organizational Identity and Access Management (IAM) systems (like Okta, Azure AD, or Google Workspace) with MCP's capability-based security model. This allows administrators to define who (user/group) can access which tools (capabilities) across the organization.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Integrate with OIDC and SAML 2.0 providers for user authentication.
    *   Map organizational groups/roles to granular MCP capability tokens.
    *   Provide a centralized policy engine for multi-tenant tool access.
    *   Support "Impersonation" where an agent acts on behalf of a specific authenticated user.
*   **Non-Goals:**
    *   Building a new identity provider (MCP Any is a consumer/bridge).
    *   Managing user passwords or PII directly.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin.
*   **Primary Goal:** Ensure that only the "Engineering" group can use the `database-admin` MCP server, while "All Employees" can use the `calendar-tool`.
*   **The Happy Path (Tasks):**
    1.  Admin configures OIDC provider in MCP Any settings.
    2.  Admin defines a mapping: `Group: Engineers` -> `Scope: db:write:*`.
    3.  A developer in the "Engineers" group authenticates via the MCP Any UI or CLI.
    4.  MCP Any issues a session-bound JWT containing the authorized MCP capabilities.
    5.  The developer's agent uses this JWT to call tools; the `Policy Firewall` verifies the scope before execution.

## 4. Design & Architecture
*   **System Flow:**
    - **Authentication**: Redirect to OIDC provider (Auth Code Flow with PKCE).
    - **Token Exchange**: Receive ID Token/Access Token.
    - **Capability Mapping**: The `IAMBridgeMiddleware` queries a mapping table (Rego/CEL) to translate provider groups into MCP scopes.
    - **Verification**: The `Policy Firewall` intercepts every `tools/call` and checks the user's mapped scopes.
*   **APIs / Interfaces:**
    - `/auth/login`, `/auth/callback` endpoints.
    - Integration with existing `Policy Firewall` middleware.
*   **Data Storage/State:** Mapping configurations stored in `config.yaml` or a dedicated encrypted DB.

## 5. Alternatives Considered
*   **Static API Keys per User**: Too hard to manage and rotate at scale; doesn't leverage existing IAM.
*   **Direct LDAP Integration**: *Rejected* in favor of modern OIDC/SAML for better compatibility with cloud IDPs.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Tokens must have short TTLs. Support for Revolving/Disposable capability tokens for high-risk tool calls.
*   **Observability:** Audit logs must include the `user_id` and `group` for every authorized tool execution.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
