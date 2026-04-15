# Design Doc: Glic-Jack Defense (GJD)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The "Glic Jack" vulnerability (CVE-2026-0628) highlights a critical weakness in agentic browser panels: they operate in a high-privilege context that can be hijacked by malicious extensions or websites via hidden prompts. Since MCP Any provides the infrastructure for browser-resident agents (via A2UI), it must actively prevent these agents from executing high-privilege actions (like camera access or local file reads) without explicit, out-of-band user attestation.

GJD is a security service that breaks the implicit trust between the browser-resident agent surface and the host OS by mandating hardware-attested approval for sensitive operations.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate hardware-attested (TPM/Secure Enclave) out-of-band approval for high-privilege tool calls initiated from browser-resident surfaces.
    * Implement real-time "Reasoning-Surface Correlation" to detect if a tool call is being driven by injected context.
    * Provide a secure, non-browser-based approval dialog (e.g., CLI prompt or native OS notification).
* **Non-Goals:**
    * Replacing standard browser security models (it supplements them).
    * Blocking all browser-initiated calls (only high-privilege/high-risk ones).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a hijacked browser panel from silently exfiltrating local files.
* **The Happy Path (Tasks):**
    1. User is using a browser-resident agent connected via MCP Any.
    2. A malicious site attempts to "Glic-Jack" the agent into calling `read_local_file`.
    3. GJD intercepts the call and identifies the source as a browser-resident surface.
    4. MCP Any pauses execution and triggers a native OS notification: "Agent in Browser requested 'read_local_file'. Approve?"
    5. User sees the unexpected request and denies it via the native prompt.
    6. The tool call is blocked, and the attempted hijack is logged.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Browser-Resident Agent] -->|Tool Call| B(A2UI Gateway)
        B --> C{GJD Middleware}
        C -->|High-Privilege?| D{Out-of-Band Approval}
        D -->|Approved| E[Execute Tool]
        D -->|Denied| F[Block & Log]
        C -->|Low-Privilege| E
    ```
* **APIs / Interfaces:**
    * `gjd.InterceptCall(surfaceID, toolCall) -> ApprovalStatus`: Evaluates risk and triggers OOB approval if needed.
    * `gjd.RegisterHighPrivilegeTools([]string)`: Defines the set of tools requiring GJD.
* **Data Storage/State:**
    * Policy registry for high-privilege tools.
    * Audit log of browser-initiated tool calls and their approval status.

## 5. Alternatives Considered
* **Purely Browser-Based Approval:** Rejected because if the browser panel is hijacked, the approval dialog itself can be manipulated or bypassed (XSS/Clickjacking).
* **Standard HITL Middleware:** Insufficient because it often relies on the same communication channel that might be compromised. GJD specifically mandates *out-of-band* (OOB) hardware-attested confirmation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** GJD enforces the principle of least privilege for non-native surfaces.
* **Observability:** Integrates with the "Origin Violation Security Hub" to report "Glic-Jack" style attempts.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
