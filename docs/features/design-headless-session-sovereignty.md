# Design Doc: Headless Session Sovereignty

**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
The introduction of "Remote Control" and "Dispatch" modes in agent frameworks like Claude Code allows agents to run as long-lived background workers. However, this creates a security gap where the primary controller (user terminal) is no longer persistently attached to the session. The "Glic Jack" vulnerability (CVE-2026-0628) also demonstrates that browser extensions can hijack agent panels if the origin is not strictly verified. Headless Session Sovereignty (HSS) ensures that background agents remain cryptographically bound to the authorized mission-root and are resilient to origin spoofing.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-attested "Sovereignty Anchor" for agents running in headless or background modes.
    * Mandate Hardware-Bound Origin Validation (HBOV) for all remote-control and re-attachment requests.
    * Implement "State Pinning" to lock agent capabilities to the specific dispatch-origin.
    * Integrate with WebView Integrity Shield (WIS) to prevent browser-extension-based hijacks.
* **Non-Goals:**
    * Multi-user real-time control (HSS focuses on a single authorized controller at a time).
    * General-purpose terminal multiplexing (this is specific to agentic control planes).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer / Remote Developer
* **Primary Goal:** Start a long-running agent task on a remote server and securely check its progress later without risk of session hijacking.
* **The Happy Path (Tasks):**
    1. The user initiates a background agent session using `mcpany dispatch --task "Refactor DB"`.
    2. MCP Any issues a hardware-attested HSS token and pins the session to the user's current hardware ID.
    3. The agent executes in the background; the user disconnects.
    4. Later, the user attempts to re-attach from a different terminal using `mcpany remote-control`.
    5. HSS performs HBOV, verifying the user's hardware identity against the session's Sovereignty Anchor.
    6. Upon successful attestation, the user is granted secure control of the background session.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Controller] -- HBOV Handshake --> B[HSS Provider]
        B -- Verify Anchor --> C[Sovereignty Vault (TPM)]
        B -- Approved --> D[Background Agent]
        D -- State Updates --> B
        B -- Origin-Locked Stream --> A
    ```
* **APIs / Interfaces:**
    * `DispatchSession(task, origin_metadata)`: Initializes a sovereign background session.
    * `ReattachSession(token, hardware_proof)`: Performs HBOV to resume control.
    * `VerifyWebViewOrigin(origin)`: Real-time check for WIS-compliant UI integration.
* **Data Storage/State:** Session anchors are stored in kernel-bound memory or a hardware-protected keystore.

## 5. Alternatives Considered
* **Password-Protected Sessions**: Rejected as insufficient against local attackers or browser-based hijacks (easily phished/sniffed).
* **IP-Based Locking**: Rejected as fragile in remote/cloud environments with dynamic IPs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: HSS enforces "No-Trust-without-Hardware" for all headless transitions.
* **Observability**: Re-attachment attempts and origin violations are logged to the "Origin Violation Security Hub."

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
