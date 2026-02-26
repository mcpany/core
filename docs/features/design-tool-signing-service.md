# Design Doc: Unified Tool Signing Service

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With Gemini CLI and Claude Code moving towards "Verified Tooling," MCP tools must now provide cryptographic proof of their origin and integrity. Many existing MCP servers lack the capability to sign their own schemas, creating a "Provenance Gap." MCP Any needs to bridge this gap by acting as a centralized "Signing Authority" for all upstream tools it manages.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically generate and attach cryptographic "Provenance Receipts" to all tool schemas served by MCP Any.
    *   Support industry-standard signing algorithms (e.g., Ed25519) and formats (e.g., JWS).
    *   Provide a secure interface for managing enterprise-grade signing keys.
    *   Integrate with external KMS (Key Management Systems) for root-of-trust.
*   **Non-Goals:**
    *   Replacing the need for upstream servers to be secure (signing only proves origin/integrity of the schema, not the tool's behavior).
    *   Acting as a general-purpose PKI (it is specific to MCP tool provenance).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Ensure all MCP tools used by the company's Gemini-based agents are verified and compliant with corporate "Verified Tooling" policies.
*   **The Happy Path (Tasks):**
    1.  Architect configures an Enterprise Root Key in MCP Any's secure vault.
    2.  MCP Any is configured to "Auto-Sign" all tools in the `production` profile.
    3.  A Gemini agent requests the tool list from MCP Any.
    4.  MCP Any intercepts the request, generates a signed provenance token for each tool, and appends it to the schema.
    5.  Gemini verifies the signature against MCP Any's public key and allows the tool call.

## 4. Design & Architecture
*   **System Flow:**
    - **Interceptor**: A middleware that hooks into the `tools/list` response.
    - **Signer**: A core component that hashes the tool schema and signs it using the configured key.
    - **Key Vault**: A secure storage layer (supporting local encrypted storage or HashiCorp Vault).
*   **APIs / Interfaces:**
    - **Schema Extension**: Adds a `_provenance` field to the standard MCP `Tool` object.
    - **Management API**: Endpoints for key rotation and status monitoring.
*   **Data Storage/State:** Public keys are exposed via a `.well-known/mcp-provenance.json` endpoint for easy verification by agents.

## 5. Alternatives Considered
*   **Upstream-Only Signing**: Requiring all tool developers to implement signing. *Rejected* as it would be too slow and would break compatibility with hundreds of existing legacy MCP servers.
*   **Manual Signing**: Requiring architects to manually sign every tool. *Rejected* due to the dynamic nature of MCP Any tool discovery.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Key access is restricted to the internal `SigningService`. Keys should ideally never leave a Hardware Security Module (HSM) if available.
*   **Observability:** Log every signing event, including the tool name, hash, and the key version used.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
