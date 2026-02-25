# Design Doc: Gateway Config Lock
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the discovery of CVE-2026-25253 in the OpenClaw ecosystem, it has become evident that local agent gateways are vulnerable to CSRF-based attacks that modify their configuration. MCP Any needs a mechanism to protect its configuration from unauthorized changes, even if a request originates from a local browser session.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent unauthorized configuration changes via CSRF or local script execution.
    * Implement an "Immutable-by-Default" state for critical tool mappings.
    * Require out-of-band (OOB) or multi-factor authentication (MFA) for configuration unlocks.
* **Non-Goals:**
    * Implementing a full user management system.
    * Protecting against physical access to the host machine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent User
* **Primary Goal:** Prevent a malicious website from disabling tool call confirmations in MCP Any.
* **The Happy Path (Tasks):**
    1. User attempts to change a tool's permission via the UI or CLI.
    2. MCP Any detects a "Locked" configuration state.
    3. MCP Any prompts the user for a "Security Key" or a confirmation on a secondary device (e.g., terminal prompt if the request came from UI).
    4. User provides the required validation.
    5. The configuration is updated and re-locked.

## 4. Design & Architecture
* **System Flow:**
    `Request -> CSRF Filter -> Config Controller -> Lock Manager -> Storage`
* **APIs / Interfaces:**
    * `POST /config/unlock`: Requires MFA/OOB token.
    * `POST /config/update`: Only succeeds if session is unlocked.
    * `GET /config/status`: Returns current lock state.
* **Data Storage/State:**
    * Lock state is stored in memory with a short TTL.
    * Persistent configuration is stored with a read-only bit on the filesystem when locked.

## 5. Alternatives Considered
* **Host-only Binding**: Rejected because CSRF can still target `localhost`.
* **Basic Auth**: Rejected because credentials can be stolen or reused across sessions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Principle of Least Privilege for the configuration process itself.
* **Observability:** All failed unlock attempts are logged with high severity.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation. Addressing CVE-2026-25253.
