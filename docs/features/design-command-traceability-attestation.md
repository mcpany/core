# Design Doc: Command Traceability Attestation
**Status:** Draft
**Created:** 2026-05-28

## 1. Context and Scope
As AI agent swarms evolve from single-task scripts to deep, hierarchical teammate meshes, the "Traceability Debt" has become a critical failure point. In current architectures (OpenClaw, AutoGen, Claude Code), commands often appear to originate from a "Ghost Identity" once they pass through multiple layers of sub-delegation. This lack of clear provenance makes it impossible to audit high-risk actions or prevent "Shadow Delegation" where a compromised subagent coerces a sibling into an unauthorized tool call.

MCP Any needs to solve this by providing a cryptographically signed "Chain of Command" that follows every instruction from the human-initiated mission root to the final tool execution. This ensures that every byte of state change and every external API call can be traced back to its authoritative source and parentage.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Issue hardware-attested "Command Origin Tokens" (COT) for every human-initiated request.
    *   Implement recursive signing for sub-delegations, creating a verifiable lineage of intent.
    *   Provide a "Traceability Middleware" that validates the COT before any high-risk tool call (e.g., shell execution, file edits).
    *   Support cross-framework traceability (Claude Code -> OpenClaw -> MCP Any).
*   **Non-Goals:**
    *   Storing the full content of every reasoning monologue (this is handled by the SRM Provider).
    *   Enforcing the *correctness* of the command (this is handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Auditor
*   **Primary Goal:** Verify that a suspicious file deletion on a production server was authorized by the root mission and not injected by a rogue subagent.
*   **The Happy Path (Tasks):**
    1.  The Auditor opens the **Traceability Dashboard** in MCP Any.
    2.  They search for the specific `fs:delete` tool call event.
    3.  The system displays the **Chain of Command**: Root Mission (Human) -> Supervisor Agent (Claude) -> File Specialist Subagent (OpenClaw).
    4.  The Auditor clicks "Verify Signatures" and the system confirms that every link in the chain is cryptographically bound to the hardware identities of the agents and the user's initial session.
    5.  The audit confirms the deletion was a valid part of the "Cleanup" intent branch.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        User->>Gateway: Initial Mission Root (Signed COT_v0)
        Gateway->>ParentAgent: Delegate Task (COT_v0)
        ParentAgent->>SubAgent: Sub-Delegate (COT_v1 = COT_v0 + Parent_Signature)
        SubAgent->>TraceabilityProvider: Tool Call Request (COT_v1)
        TraceabilityProvider->>PolicyEngine: Validate Lineage (COT_v1)
        PolicyEngine-->>TraceabilityProvider: Authorized
        TraceabilityProvider->>MCP_Server: Execute Tool
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/trace/sign`: Accepts a parent COT and returns a recursively signed child COT.
    *   `GET /v1/trace/verify`: Validates the complete lineage of a COT.
*   **Data Storage/State:** COTs are ephemeral and passed in headers (`x-mcp-command-trace`), but their roots are pinned in the Blackboard for the duration of the mission.

## 5. Alternatives Considered
*   **Centralized Logging:** Rejected because logs can be tampered with in a compromised environment. Cryptographic chains provide non-repudiable proof.
*   **Simple ID Passing:** Rejected because it doesn't prevent "Identity Spoofing" where a subagent claims to be its parent.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** COTs must be hardware-bound (TPM/Secure Enclave) to prevent extraction and reuse.
*   **Observability:** Every trace verification failure triggers a P0 Security Alert in the UI.

## 7. Evolutionary Changelog
*   **2026-05-28:** Initial Document Creation.
