# Design Doc: Link-Preview Interdiction Middleware

**Status:** Draft
**Created:** 2026-03-30

## 1. Context and Scope
Recent security findings have highlighted a critical data exfiltration vector where AI agents, when posting links in messaging platforms (e.g., Telegram, Discord, Slack), trigger automated "Link Previews." These previews can be weaponized by attackers to exfiltrate sensitive data contained in URL parameters or to confirm the agent's internal state without any user interaction. As MCP Any evolves into a universal gateway, it must provide a "Semantic Sovereignty" layer that interdicts and sanitizes these outputs.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect and sanitize URLs in agent-generated outputs before they reach external platforms.
    * Provide a configurable "Link-Preview Policy" (e.g., `strip`, `obfuscate`, `user-approval-required`).
    * Integrate with the "Semantic Integrity Bridge" to detect intent-based exfiltration attempts via URLs.
* **Non-Goals:**
    * Blocking all external links (which would break legitimate agent functionality).
    * Modifying the external messaging platforms' preview logic directly.

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Officer
* **Primary Goal:** Prevent a specialized research agent from accidentally exfiltrating internal project IDs via link previews in a shared Slack channel.
* **The Happy Path (Tasks):**
    1. The agent generates a message containing a link to an internal resource.
    2. The message passes through the MCP Any Link-Preview Interdiction Middleware.
    3. The middleware identifies the link and checks it against the "Sensitive Domain List."
    4. Based on the `strip-params` policy, the middleware removes high-entropy query parameters.
    5. The sanitized message is forwarded to the messaging platform, preventing an automated exfiltration preview.

## 4. Design & Architecture
* **System Flow:**
    * Agent Output -> **Link-Preview Interdiction Middleware** -> External Adapter -> Messaging Platform.
    * The middleware uses regex and semantic analysis to identify URLs and their potential for exfiltration.
* **APIs / Interfaces:**
    * `SanitizeOutput(message_body)`: Scans and modifies URLs based on the active security policy.
    * `RegisterSensitiveDomain(domain_pattern, policy)`: Configures domain-specific interdiction rules.
* **Data Storage/State:** Policies are stored in the global "Policy Engine"; sensitive domain lists are hardware-attested.

## 5. Alternatives Considered
* **Disabling Link Previews in Client Apps**: Rejected because it requires manual configuration on every client (e.g., Telegram's `linkPreview: false`) and cannot be enforced centrally by the agent infrastructure.
* **Domain Allow-listing**: Considered but rejected as a standalone solution, as even allow-listed domains can be used for exfiltration via query parameters.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Part of the "Zero-Trust Mesh" strategy, ensuring that output sovereignty is maintained regardless of the destination platform.
* **Observability**: Interdicted links are logged as "Security Events" for audit trails.

## 7. Evolutionary Changelog
* **2026-03-30:** Initial Document Creation.
