# Design Doc: MCP Security Auditor Tool
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
With the rapid expansion of the MCP ecosystem, many servers are being deployed with insecure defaults, such as unauthenticated admin panels exposed on `0.0.0.0`. Recent reports indicate over 8,000 such servers are currently reachable on the public internet. MCP Any needs a proactive mechanism to protect users by auditing their connected upstreams and ensuring they meet a minimum security baseline.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically scan all configured MCP upstreams for common misconfigurations.
    * Detect exposed admin/debug endpoints.
    * Verify the presence of authentication (tokens/keys).
    * Warn users about "0.0.0.0" bindings for local tools.
    * Provide actionable remediation steps.
* **Non-Goals:**
    * Performing deep penetration testing or fuzzing.
    * Fixing vulnerabilities automatically (outside of MCP Any's own config).
    * Auditing non-MCP services.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Ensure that the 5+ MCP servers they've connected to MCP Any are not leaking secrets or exposed to the internet.
* **The Happy Path (Tasks):**
    1. User connects a new MCP server (e.g., a local Postgres MCP) via the UI or config.
    2. MCP Any automatically triggers a background "Security Audit" for the new service.
    3. The auditor detects that the Postgres MCP admin panel is accessible on port 8080 without auth.
    4. MCP Any displays a high-priority warning in the Dashboard with a "Fix: Bind to Localhost" suggestion.
    5. User applies the fix, and the Auditor confirms the service is now "Secure."

## 4. Design & Architecture
* **System Flow:**
    * **Audit Manager**: Orchestrates the scanning lifecycle when a service is added or reloaded.
    * **Probe Engine**: Executes specific security probes (e.g., `GET /admin`, `GET /debug`, `OPTIONS *`).
    * **Heuristics Library**: A set of Rego-based rules to classify probe results (e.g., "Status 200 on /admin with no Auth header = CRITICAL").
* **APIs / Interfaces:**
    * `GET /api/v1/services/{id}/audit`: Returns the latest audit report for a service.
    * `POST /api/v1/audit/scan-all`: Manually triggers a fleet-wide scan.
* **Data Storage/State:**
    * Audit results are stored in the internal SQLite database linked to the service record.

## 5. Alternatives Considered
* **Manual Scanning**: Rejected as it places too much burden on the user and doesn't scale with agent autonomy.
* **External Security Scanners (e.g., Nuclei)**: Rejected because we want a lightweight, protocol-aware tool integrated directly into the MCP Any lifecycle without external dependencies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The auditor itself must run with minimal privileges and not transmit sensitive findings to external telemetry.
* **Observability**: Audit failures should trigger alerts in the Service Health History timeline.

## 7. Evolutionary Changelog
* **2026-02-26**: Initial Document Creation.
