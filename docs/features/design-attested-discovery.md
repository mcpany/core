# Design Doc: Attested Discovery & Runtime Integrity

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
With the rapid proliferation of MCP servers and the emergence of "personal AI agents" like OpenClaw, the attack surface for AI tools has expanded. Recent vulnerabilities (CVE-2026-23744, CVE-2026-27735) have demonstrated that simple source verification is insufficient. Attested Discovery ensures that not only is the source of the tool known, but its execution environment (sandbox or local host) is verified to be in a secure state before any tool execution occurs.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement cryptographic verification of MCP server provenance (where the tool comes from).
    *   Establish a "Runtime Attestation" protocol that checks the integrity of the execution environment.
    *   Integrate with the CEL-based Policy Engine to allow/deny tool execution based on attestation results.
    *   Provide a "Safety Score" for tools based on their attestation history.
*   **Non-Goals:**
    *   Providing the sandbox itself (MCP Any integrates with existing sandboxes like Docker or Kata).
    *   Replacing traditional OS-level security.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Enterprise Admin.
*   **Primary Goal:** Ensure that a newly discovered OpenClaw tool for "Email Management" cannot execute if its sandbox environment has been tampered with.
*   **The Happy Path (Tasks):**
    1.  User adds a new MCP server URL to MCP Any.
    2.  MCP Any performs a "Discovery Handshake" that includes a request for a Cryptographic Provenance Token.
    3.  The MCP server provides the token along with a signed Runtime Attestation Report (e.g., via TPM or Secure Boot measurements).
    4.  MCP Any verifies the signatures against a trusted registry.
    5.  The Policy Engine evaluates the report: `policy: "allow if attestation.integrity_score > 0.9"`.
    6.  The tool is registered and made available to the agent.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery Phase**: `DiscoveryMiddleware` intercepts the `list_tools` call and requests attestation.
    - **Verification Phase**: The `AttestationProvider` validates the provided signatures and reports.
    - **Enforcement Phase**: The `PolicyFirewall` (CEL-based) uses the attestation metadata to make the final decision.
*   **APIs / Interfaces:**
    - New `mcp_attestation_v1` protocol extension for the handshake.
    - Integration with existing MCP `initialize` and `list_tools` methods.
*   **Data Storage/State:** Attestation certificates and trusted root hashes are stored in the secure configuration vault.

## 5. Alternatives Considered
*   **Static Whitelisting**: Only allowing tools from pre-approved domains. *Rejected* as it doesn't account for environment tampering or compromised domains.
*   **Post-Execution Auditing**: Detecting breaches after they happen. *Rejected* as it doesn't prevent the initial exploit.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust feature. It moves from "Trusting the Source" to "Verifying the State."
*   **Observability:** The UI will display a "Shield" icon next to attested tools, with a detailed breakdown of the integrity report.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
