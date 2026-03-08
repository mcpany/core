# Design Doc: Anti-Smuggling Context Filter

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
The "Context Smuggling" vulnerability (discovered 2026-03-04) allows malicious tool outputs to inject specialized markers into MCP result payloads. When these payloads are processed by recursive context middleware, they can "smuggle" unauthorized instructions or context into a subagent's execution environment. This bypasses the intent-scoping of the `Recursive Context Protocol` (RCP). The Anti-Smuggling Context Filter is a high-priority middleware designed to sanitize and attest to the integrity of all context blocks.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement cryptographic signing (Ed25519) for all parent-to-child context blocks.
    *   Verify the "Lineage" of every context entry before it's injected into a tool call.
    *   Provide an "Integrity-First" filtering layer that strips unverified or suspicious markers from tool outputs.
    *   Enable "Attested Context" headers in all inter-agent communications.
*   **Non-Goals:**
    *   Completely redesigning the existing `Recursive Context Protocol`.
    *   Solving all forms of prompt injection (this focuses specifically on "smuggling" via tool results).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Architect.
*   **Primary Goal:** Prevent a subagent from being hijacked by a malicious tool result that "smuggles" a new system prompt into its context.
*   **The Happy Path (Tasks):**
    1.  Architect enables `integrity_attestation: true` in the MCP Any config.
    2.  Parent agent initializes a context session; MCP Any signs the initial state.
    3.  A subagent calls a "Search" tool which returns a malicious payload containing a "smuggled" instruction.
    4.  The Anti-Smuggling Filter detects the unverified instruction markers in the tool output.
    5.  The filter strips the smuggled content and logs a "Security: Context Smuggling Attempt Blocked" event.
    6.  The subagent receives only the sanitized tool output.

## 4. Design & Architecture
*   **System Flow:**
    - **Context Signature**: Every entry in the `Shared KV Store` (Blackboard) is stored with an Ed25519 signature of its `(key, value, parent_id, timestamp)`.
    - **Middleware Hook**: The `AntiSmugglingMiddleware` intercepts all `tools/call` results. It scans for known "Smuggling Markers" (e.g., `[[MCP_CONTEXT_OVERRIDE]]`) and validates that any such markers are backed by a valid signature from a trusted internal authority.
*   **APIs / Interfaces:**
    - Metadata extension for context entries: `_mcp_context_signature: "<ed25519_sig>"`
    - Policy Engine integration: `allow_unsigned_context: false`
*   **Data Storage/State:** Storage of the instance's private signing key in a secure, local-only enclave (e.g., `~/.mcpany/enclave.key`).

## 5. Alternatives Considered
*   **Simple Regex Filtering**: Block specific keywords. *Rejected* because smuggling markers can be obfuscated or encoded.
*   **Full Context Re-Verification**: Re-verify the entire context on every tool call. *Rejected* due to excessive latency; signature-based attestation is more efficient.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a critical component of the Zero Trust architecture, ensuring that context itself cannot be used as an attack vector.
*   **Observability:** The UI "Security Dashboard" will show a real-time counter of "Smuggling Attempts Blocked."

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
