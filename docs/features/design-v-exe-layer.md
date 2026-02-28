# Design Doc: Verifiable Execution (V-EXE) Layer

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The February 2026 "OpenClaw Security Crisis" and the "ClawHavoc" incident (335+ malicious skills distributed) demonstrated that current agent architectures are vulnerable to "Shadow Skill" injection. In these attacks, a seemingly benign tool call can trigger unauthorized remote code execution or data exfiltration because the agent (and the gateway) cannot verify the authenticity of the tool's implementation at runtime. MCP Any needs a Verifiable Execution (V-EXE) layer that ensures every tool execution is cryptographically bound to an attested source.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a "Challenge-Response" handshake between MCP Any and the Tool Provider (Upstream).
    *   Verify the cryptographic signature (Ed25519/RSA) of the tool definition and the binary/script being executed.
    *   Integrate with the `Policy Firewall` to block any tool that fails attestation.
    *   Provide a "Provenance Ledger" for auditing tool execution history.
*   **Non-Goals:**
    *   Providing a sandboxed environment (this is handled by the OS/Container layer).
    *   Verifying the *logic* of the code (V-EXE focuses on *provenance* and *integrity*).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Prevent "Shadow Skills" from executing on developer machines.
*   **The Happy Path (Tasks):**
    1.  The Architect configures a "V-EXE Required" policy in MCP Any.
    2.  An AI Agent attempts to call `execute_shell_command` from a new community MCP server.
    3.  The V-EXE Layer intercepts the call and requests a signature from the MCP server.
    4.  The MCP server provides a signature bound to its registered developer ID.
    5.  MCP Any verifies the signature against its trusted CA/Registry.
    6.  If valid, execution proceeds; otherwise, it is blocked and an alert is raised.

## 4. Design & Architecture
*   **System Flow:**
    - **Manifest Verification**: On tool discovery, MCP Any validates the `mcp_manifest.sig`.
    - **Runtime Handshake**: For P0 (Privileged) tools, MCP Any sends a nonce to the tool provider. The provider must return the signature of `nonce + request_body`.
    - **Attestation Service**: A lightweight internal service that manages trusted public keys.
*   **APIs / Interfaces:**
    - Extension to MCP `callTool`: `_v_exe_attestation: { signature: string, nonce: string }`.
    - Internal `PolicyEngine` hook: `OnPreExecute(tool, request)`.
*   **Data Storage/State:** Trusted keys are stored in the `Service Registry`. Nonces are transient and stored in memory.

## 5. Alternatives Considered
*   **Hash-based Verification (Static)**: Only verifying the hash of the configuration. *Rejected* because it doesn't protect against dynamic supply-chain poisoning where the endpoint remains the same but the implementation changes.
*   **Manual Approval (HITL)**: Asking the user for every call. *Rejected* as it's not scalable for autonomous agents.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust feature. It moves security from "Trust on First Use" (TOFU) to "Always Verify."
*   **Observability:** The UI must show "Attestation Status" for every tool call in the trace.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
