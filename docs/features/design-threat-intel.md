# Design Doc: Real-time Threat Intel Middleware

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The "ClawHavoc" campaign demonstrated that the AI agent ecosystem is vulnerable to large-scale supply chain poisoning. Malicious skills and MCP servers can be injected into marketplaces, leading to credential theft (e.g., AMOS stealer) and RCE. MCP Any, as a universal gateway, is uniquely positioned to act as a "firewall" that blocks known malicious tools before they are even presented to an LLM or executed.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Integrate with external Threat Intelligence feeds (hashes of malicious skills, known bad MCP server origins).
    *   Automatically quarantine tools that match "High Risk" signatures.
    *   Provide real-time alerts in the UI when a poisoned tool is detected.
    *   Support "Community Attestation" where users can report suspicious tool behavior.
*   **Non-Goals:**
    *   Performing deep behavioral analysis of arbitrary code (this is handled by the "Sandboxed Execution" feature).
    *   Acting as a general-purpose Antivirus for the host OS.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin.
*   **Primary Goal:** Prevent employees from accidentally connecting a poisoned "OpenClaw" skill to the corporate MCP gateway.
*   **The Happy Path (Tasks):**
    1.  Admin enables "Threat Intel Feed" in MCP Any settings.
    2.  An employee attempts to register a new MCP server that was recently flagged in the ClawHavoc campaign.
    3.  The Threat Intel Middleware intercepts the registration.
    4.  The registration is blocked, and a security alert is generated: "Registration Blocked: Known Malicious Source (ClawHavoc-2026-03-05)."
    5.  The admin receives a notification in the Security Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    - **Feed Sync**: A background service periodically pulls signature updates (SHA256 hashes, URL patterns) from verified providers.
    - **Interception**: During `tools/list` or service registration, the middleware computes hashes of the tool definitions and checks them against the local threat database.
    - **Quarantine**: If a match is found, the tool is marked as `STATUS_QUARANTINED` and hidden from the LLM.
*   **APIs / Interfaces:**
    - Internal `ThreatScanner` interface.
    - `GET /api/security/threats`: List detected and blocked threats.
*   **Data Storage/State:** Local SQLite database for cached threat signatures.

## 5. Alternatives Considered
*   **Manual Vetting Only**: Relying on admins to manually check every tool. *Rejected* due to the speed and volume of modern agentic supply chain attacks.
*   **Centralized Proxying**: Routing all tool calls through a central security cloud. *Rejected* to preserve the "Local-First" and privacy goals of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The middleware itself must be protected from "Feed Poisoning" using cryptographic signatures on the threat intelligence updates.
*   **Observability:** Integrated with the "Supply Chain Attestation Viewer" in the UI.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
