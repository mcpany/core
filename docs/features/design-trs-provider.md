# Design Doc: Temporal Remote Sovereignty (TRS) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The rapid adoption of "Claude Code Remote Control" and Gemini "Remote Execution" has introduced a significant threat vector: **Persistence without Presence**. A persistent remote session (Remote-as-Local) allows an attacker who has compromised a developer's cloud account to pivot directly to their local host environment. Auto-reconnect features mean that "a forgotten open terminal is an open door."

TRS addresses this by mandating hardware-bound "Presence Heartbeats." It ensures that a remote session is only sovereign as long as the authorized developer is physically present at the hardware root.

## 2. Goals & Non-Goals
* **Goals:**
    * Broker time-bound, hardware-attested session tokens for all remote agent interactions.
    * Automatically revoke capabilities if a "Presence Heartbeat" (via TPM or Biometric) is missed.
    * Implement "Remote-to-Local" isolation boundaries that prevent lateral movement beyond the active project.
    * Provide a unified "Presence Status" to the entire agent mesh.
* **Non-Goals:**
    * Managing user login/SSO (handled by Auth0/Okta).
    * Providing the remote tunneling itself (handled by AMT).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Enable remote control for a long-running code migration mission while ensuring host safety if they leave their desk.
* **The Happy Path (Tasks):**
    1. Developer activates "TRS Mode" on their local MCP Any instance.
    2. MCP Any requests a `Presence Attestation` from the local TPM.
    3. TRS issues a 15-minute `Sovereignty Lease`.
    4. Remote agent executes migration tools.
    5. Developer locks their screen and walks away.
    6. TPM stops broadcasting the "Presence" signal.
    7. After 15 minutes, the Sovereignty Lease expires.
    8. MCP Any atomically revokes all AMT tunnels and locks the project-local configurations.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Local TPM/Biometric] -- Pulse --> B[TRS Provider]
        B -- Lease Renewal --> C[AMT Bridge]
        C -- Sovereign Link --> D[Remote Agent]
        B -- Expiry Signal --> E[Capability Revoker]
        E -- Lock --> F[Host Tools]
    ```
* **APIs / Interfaces:**
    * `GET /v1/sovereignty/pulse`: Local hardware-bound presence check.
    * `POST /v1/sovereignty/revoke`: Forceful revocation of remote leases.
* **Data Storage/State:**
    * Sovereignty leases are stored in `memfd` regions with automatic kernel-level purging upon expiry.

## 5. Alternatives Considered
* **Short-Lived JWTs**: Rejected as they can be re-played if the transport is compromised. TRS requires a hardware-bound *pulse*.
* **Manual Revocation**: Rejected as humans are the primary failure point in "Persistent Terminal" exploits.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Pulse" must include a monotonic counter from the TPM to prevent replay of the presence signal.
* **Observability:** Presence status and lease expiries are visualized in the **Remote Presence Heartbeat Widget**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
