# Design Doc: Side-Channel Interdiction Middleware (SCIM)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The disclosure of CVE-2026-0628 highlights a critical vulnerability where low-privilege components (like browser extensions) can hijack elevated AI capabilities by bridging local communication gaps. In multi-agent swarms, specialist agents often operate at different trust levels. SCIM is required to prevent "Side-Channel" communication—such as probing local ports, using unauthorized named pipes, or exploiting shared metadata—to ensure that inter-agent coordination is strictly confined to the verified, hardware-attested mainline transport.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor and intercept all out-of-band communication attempts between subagents.
    * Enforce "MAINLINE-ONLY" coordination policies.
    * Detect "EchoLeak" patterns where secrets are leaked via timing or metadata side-channels.
* **Non-Goals:**
    * Encrypting all traffic (handled by the T2T Encryption Bridge).
    * Policing user-initiated local network traffic outside the agent scope.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor
* **Primary Goal:** Ensure a low-trust "Web Searcher" agent cannot communicate with a high-trust "Database Admin" agent except through the audited Mission-Root mailbox.
* **The Happy Path (Tasks):**
    1. SCIM initializes as part of the T2T coordination bus.
    2. The "Web Searcher" attempt to open a direct socket to a local port known to be used by the "Database Admin."
    3. SCIM's kernel-level monitor detects the unauthorized probe.
    4. SCIM interdicts the connection and issues a "Side-Channel Violation" alert.
    5. The event is logged with full stylometric and environmental context for the Auditor.

## 4. Design & Architecture
* **System Flow:**
    `Subagent A` --(Unauthorized Probe)--> `[SCIM Monitor]` --(Interdiction)--> `Block`
* **APIs / Interfaces:**
    * Internal hook into the `Isolated Named-Pipe Transport`.
    * `GET /v1/security/side-channels`: Returns active interdiction events.
* **Data Storage/State:**
    * Maintains a "Trust Mesh" map of authorized communication paths.

## 5. Alternatives Considered
* **Strict Network Namespacing (NS)**: Rejected as the primary solution because it doesn't catch metadata-layer "EchoLeaks" within the same namespace. SCIM provides semantic-level interdiction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SCIM assumes all non-mainline paths are malicious.
* **Observability:** Integrates with the `CSAD Hub` for collective anomaly detection.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
