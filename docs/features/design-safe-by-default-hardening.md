# Design Doc: Safe-by-Default Infrastructure Hardening

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The February 2026 security crisis (8,000+ exposed MCP servers, Clawdbot breach) highlighted a critical failure in the agentic ecosystem: ease-of-use was prioritized over security. Many users unknowingly bind MCP gateways to `0.0.0.0`, exposing sensitive tools and environment variables to the public internet. MCP Any must transition to a "Safe-by-Default" posture where the system is inherently secure even for novice users.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce `localhost` (`127.0.0.1` / `::1`) bindings for all adapters and gateways by default.
    *   Implement a "Remote Access Guard" that prevents `0.0.0.0` or non-local bindings without explicit administrative attestation.
    *   Introduce cryptographic MFA/Attestation for any remote management or tool access.
    *   Provide automated "Exposure Check" on startup.
*   **Non-Goals:**
    *   Completely disabling remote access (it must remain an option for enterprise use).
    *   Managing host-level firewall rules (MCP Any should focus on its own listener configuration).

## 3. Critical User Journey (CUJ)
*   **User Persona:** New AI Engineer deploying MCP Any for the first time.
*   **Primary Goal:** Set up the gateway without accidentally exposing tools to the internet.
*   **The Happy Path (Tasks):**
    1.  User runs `mcpany start` without a configuration file.
    2.  MCP Any binds all services to `127.0.0.1` and outputs a "Secure Local Mode" banner.
    3.  If the user attempts to set `host: 0.0.0.0` in the config, the server fails to start with a "Security Override Required" error.
    4.  User follows instructions to generate an `access_attestation.token` to enable remote exposure.

## 4. Design & Architecture
*   **System Flow:**
    - **Listener Configuration**: The `ConfigLoader` validates the `host` parameter. If non-local, it checks for a valid `AttestationToken`.
    - **Security Bootstrap**: On first run, a unique cryptographic identity (Ed25519) is generated for the instance.
    - **MFA Layer**: Remote access requests must include a signature from the instance's private key, typically handled via a "Second Screen" approval on the local machine.
*   **APIs / Interfaces:**
    - New CLI command: `mcpany secure authorize-remote [ip]`
    - Metadata extension for tool calls: `_mcp_source_locality: "local" | "remote"`
*   **Data Storage/State:** Secure storage of the instance identity in a protected file (e.g., `~/.mcpany/id_ed25519`).

## 5. Alternatives Considered
*   **Just Adding a Warning**: Log a warning when binding to `0.0.0.0`. *Rejected* as history shows users ignore logs.
*   **Requiring Docker Networking**: Forcing users to use Docker to isolate ports. *Rejected* as it adds too much friction for non-Docker workflows.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature is the foundation of the Zero Trust architecture. It ensures that the first point of failure (misconfiguration) is mitigated.
*   **Observability:** The UI should prominently display "Connectivity Status: [Local Only | Remote Authorized]" with a list of active remote listeners.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.

### Update: 2026-03-02 - Mitigating "MCP Top 10" Risks
**Context:** Today's market sync highlighted the formalization of the "MCP Top 10" vulnerabilities, specifically "Unauthenticated Discovery" and "Tool Poisoning."
**Architecture Adjustment:**
*   **Mandatory Local-First Bindings**: All new MCP Any instances will now bind exclusively to `127.0.0.1`. Remote exposure via `0.0.0.0` will require an explicit `ALLOW_REMOTE_EXPOSURE=true` flag AND a valid Attestation Token.
*   **Discovery Sanitization Middleware**: Introduced a "Sanitization Hook" that runs before any tool schema is returned. This hook automatically strips potentially malicious script tags or unexpected fields from the tool definition.
**Security Impact:** Prevents accidental exposure to the public internet and mitigates common "Tool Poisoning" vectors found in the current MCP registry.
