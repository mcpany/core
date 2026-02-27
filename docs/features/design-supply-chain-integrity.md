# Design Doc: Supply Chain Integrity Guard (MCP Attestation)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The "ClawHavoc" campaign and the earlier "Clinejection" attacks have demonstrated that the greatest threat to AI agent security is the injection of malicious tools or skills. As agents become more autonomous, they increasingly fetch and install tools from third-party repositories. MCP Any must ensure that every MCP server it connects to is verified and has not been tampered with.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Verify the cryptographic provenance of MCP server binaries and configuration files.
    *   Implement a "Signed Tooling" requirement for all non-local MCP servers.
    *   Provide a centralized registry of "Trusted Publishers" for MCP tools.
    *   Integrate with the Policy Firewall to block unverified tool registration.
*   **Non-Goals:**
    *   Developing a full Public Key Infrastructure (PKI) - we will leverage existing standards like Cosign or Sigstore.
    *   Sandboxing the tool execution itself (this is handled by the transport layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Administrator.
*   **Primary Goal:** Prevent the accidental registration of a typosquatted, malicious MCP tool from a public repository.
*   **The Happy Path (Tasks):**
    1.  Admin configures MCP Any to "Strict Attestation Mode."
    2.  An agent attempts to register a new tool: `github.com/mcp-tools/sql-helper`.
    3.  The `AttestationMiddleware` fetches the cryptographic signature for the tool.
    4.  The signature is validated against the "Trusted Publishers" list.
    5.  If validation fails, the registration is rejected, and an alert is logged in the Security Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    - **Registration Hook**: The `MCPDiscoveryService` calls the `AttestationMiddleware` during the registration phase.
    - **Verification Loop**: The middleware checks for a `.sig` or `.attestation` file accompanying the tool configuration.
    - **Policy Check**: The Policy Firewall verifies the publisher's identity against allowed OIDC identities (e.g., GitHub Actions, verified developer emails).
*   **APIs / Interfaces:**
    - `POST /api/v1/attestations/verify`: Endpoint for manual verification of a tool configuration.
    - `GET /api/v1/security/trusted-publishers`: List of verified tool publishers.
*   **Data Storage/State:** Signatures and trust metadata are stored in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Manual Approval (HITL)**: Requiring a human to approve every tool. *Rejected* as too slow for dynamic agent swarms, though it remains a fallback.
*   **Runtime Sandboxing Only**: Allowing any tool but restricting its filesystem/network access. *Rejected* because malicious tools can still exfiltrate data via allowed channels or engage in prompt injection.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core security feature. We must ensure the Attestation Registry itself is tamper-proof.
*   **Observability:** The UI must clearly display the "Attestation Status" (Verified/Unsigned/Failed) for every tool in the library.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
