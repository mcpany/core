# Design Doc: MCP Provenance Attestation

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The rapid expansion of the MCP ecosystem has led to the emergence of "shadow MCP servers" and untrusted tool sources. Recent supply chain attacks (e.g., "Clinejection") demonstrate that agents can be tricked into installing or calling malicious MCP servers that look legitimate. MCP Provenance Attestation is needed to ensure that every MCP server connected to MCP Any is cryptographically verified and originates from a trusted source.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a cryptographic signature verification mechanism for MCP server manifests.
    * Provide a "Trusted Registry" integration where MCP Any can verify server identities.
    * Enable "Strict Attestation Mode" where unverified servers are blocked.
    * Support decentralized attestation via DIDs (Decentralized Identifiers).
* **Non-Goals:**
    * Building a centralized App Store for MCP (MCP Any remains decentralized).
    * Auditing the *code* of the MCP server (this is handled by AitM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Platform Engineer.
* **Primary Goal:** Prevent an AI agent from accidentally calling a rogue "Calculator" tool that actually exfiltrates environment variables.
* **The Happy Path (Tasks):**
    1. Engineer enables `Strict Attestation` in the MCP Any Policy Engine.
    2. An AI agent attempts to add a new MCP server from a URL.
    3. MCP Any fetches the server's `manifest.json` and a corresponding `provenance.sig`.
    4. MCP Any verifies the signature against a trusted root of trust or a known DID.
    5. The server is either admitted to the mesh or rejected with a security warning.

## 4. Design & Architecture
* **System Flow:**
    - **Manifest Fetching**: When a service is registered, MCP Any looks for attestation metadata.
    - **Verification Pipeline**: A new `ProvenanceVerificationMiddleware` intercepts service registration. It uses public keys (local or via OIDC/DID) to verify signatures.
    - **Attestation Tokens**: Verified services receive a "Provenance Token" stored in the `Service Registry`.
* **APIs / Interfaces:**
    - **Internal**: `VerifyProvenance(manifest []byte, signature []byte) error`.
    - **Config**: New `attestation` block in service configuration.
* **Data Storage/State:** Public keys and trust roots are stored in the SQLite database under a new `trust_roots` table.

## 5. Alternatives Considered
* **Manual Whitelisting**: Rejected due to lack of scalability in dynamic agent swarms.
* **Hash-based Pinning**: Effective but doesn't handle updates well. Combined with signatures for best results.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a foundation for Zero Trust. Without provenance, capability tokens can be forged or stolen by rogue upstreams.
* **Observability:** Verification status is displayed in the Service List UI and logged in the Security Audit log.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
