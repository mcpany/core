# Design Doc: Runtime Attestation Manifest (RAM)
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As AI agents like Gemini CLI and Claude Code move towards "Verified Tooling," MCP servers can no longer be anonymous providers. Clients need a way to verify that an MCP server is who it claims to be, is running the code it claims to run, and that its tool definitions haven't been tampered with. The "Ghost-Tool" exploit further emphasizes the need for a verifiable source of truth for tool schemas and origins.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographically signed manifest of all registered tools.
    * Enable external clients to verify the server's binary integrity (via build-time hashing).
    * Include environment metadata (e.g., allowed egress domains) in the manifest.
* **Non-Goals:**
    * Encrypting tool responses (this is handled by transport layer security).
    * Managing client-side identities (this is handled by MFA/Auth middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator (e.g., Gemini CLI)
* **Primary Goal:** Verify the integrity and safety of MCP Any before allowing it to execute local shell commands.
* **The Happy Path (Tasks):**
    1. Client connects to MCP Any and requests the `/.well-known/mcp-manifest.json`.
    2. MCP Any generates a manifest containing tool hashes, server version, and a cryptographic signature.
    3. Client verifies the signature against a trusted public key (or a decentralized registry).
    4. Client checks if the tool it wants to call matches the hash in the manifest.
    5. Client proceeds with the tool call, confident that the tool hasn't been "poisoned."

## 4. Design & Architecture
* **System Flow:**
    - At startup, the `ServiceRegistry` calculates a SHA-256 hash for every registered tool schema and its associated metadata.
    - The `AttestationProvider` collects these hashes, along with server build info (Commit SHA, Build Date).
    - A signing key (configured via `MCPANY_ATTESTATION_KEY`) is used to sign the manifest.
    - The manifest is served via a new internal MCP tool `get_runtime_manifest` or a dedicated HTTP endpoint.
* **APIs / Interfaces:**
    - `GetManifest()` returns the `RuntimeManifest` struct.
    - `VerifyManifest(m RuntimeManifest) error` (for client-side use).
* **Data Storage/State:**
    - The manifest is generated and cached in memory at startup. Re-generated only on configuration hot-reloads.

## 5. Alternatives Considered
* **On-the-fly Signing:** Rejected because it's computationally expensive and adds latency to tool discovery.
* **Centralized Registry:** Considered but rejected for the initial phase to maintain MCP Any's "Local-First" philosophy. Decentralized verification can be added later.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The signing key must be protected. If the key is compromised, the manifest can be spoofed.
* **Observability:** Logs will record whenever a manifest is requested and if signature verification fails on the client-side.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
