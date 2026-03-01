# Design Doc: Mandatory Tool Provenance Signing
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The "OpenClaw Security Crisis" (Feb 2026) demonstrated that autonomous agents are highly vulnerable to "Skill Poisoning" and RCE via unvetted third-party tools. MCP Any, as a universal gateway, must ensure that every tool it proxies comes from a verified and untampered source.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a cryptographic verification layer for all MCP server configurations.
    *   Support "Attestation Reports" for local and remote MCP servers.
    *   Provide a "Zero-Trust" execution sandbox for unverified tools.
    *   Enable community-based reputation scoring integrated into the provenance check.
*   **Non-Goals:**
    *   Providing a certificate authority (we will leverage existing OIDC/Sigstore infrastructure).
    *   Verifying the *logic* of the code (this is about *origin* and *integrity*).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious DevOps Lead.
*   **Primary Goal:** Prevent the execution of unauthorized or tampered tools in the production agent swarm.
*   **The Happy Path (Tasks):**
    1.  Admin adds a new MCP server to MCP Any via a signed manifest.
    2.  MCP Any validates the signature against a trusted public key (e.g., via GitHub Actions OIDC or Sigstore).
    3.  During tool execution, MCP Any checks that the server's binary/image hash matches the signed manifest.
    4.  If verification fails, the tool call is blocked or routed to an isolated, network-gapped sandbox.

## 4. Design & Architecture
*   **System Flow:**
    - **Manifest Verification**: MCP servers must include an `mcp-manifest.json` signed with a valid developer key.
    - **Provenance Guard**: A middleware that intercepts every `tools/call` and verifies the server's runtime integrity (PID/Image hash) before forwarding the request.
    - **Trust Levels**: Tools are categorized as `Verified`, `Community-Trusted`, or `Untrusted (Sandboxed)`.
*   **APIs / Interfaces:**
    - New Admin API: `/admin/provenance/verify`
    - Policy Engine Integration: `allow if input.tool.provenance == "Verified"`
*   **Data Storage/State:** Local cache of verified hashes and trusted public keys.

## 5. Alternatives Considered
*   **Runtime Sandboxing Only**: Just run everything in a sandbox. *Rejected* as too restrictive for many legitimate local tools (e.g., filesystem access) and high performance overhead.
*   **Manual Review**: Humans review every tool. *Rejected* as it doesn't scale with the thousands of tools being generated.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The core of this design. It moves from "trust by default" to "verify then trust."
*   **Observability:** All provenance failures must be logged with high-fidelity telemetry for incident response.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation following the OpenClaw/ClawHub security incidents.
