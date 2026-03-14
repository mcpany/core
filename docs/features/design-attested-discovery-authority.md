# Design Doc: Attested Discovery Authority
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
As the AI agent ecosystem matures, specifically with the introduction of hardened MCP trust in Claude Code (addressing CVE-2025-59536), there is a critical need for a centralized, reliable method to verify the identity and provenance of local MCP servers. MCP Any must transition from a simple tool proxy to an authoritative trust broker. This feature implements the "Attested Discovery Authority" (ADA) to provide cryptographic proof of a tool's provenance before it is exposed to the agent runtime.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographic identity broker for local MCP servers.
    * Enable agent frameworks (like Claude Code) to satisfy "Trust Verification" requirements.
    * Support multi-signature attestation from both the MCP server and the user's security policy.
    * Integrate with hardware-bound (TPM/SEP) keys for root-of-trust.
* **Non-Goals:**
    * Replacing existing MCP transport protocols.
    * Providing a global, centralized registry for all MCP servers (focus is on local/enterprise trust).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer using Claude Code.
* **Primary Goal:** Safely use a new, locally-developed MCP server for code analysis without triggering security warnings or risking RCE.
* **The Happy Path (Tasks):**
    1. User starts MCP Any and a new local MCP server.
    2. MCP Any detects the new server via its auto-discovery providers.
    3. The local MCP server presents its identity claim (e.g., a signed manifest).
    4. MCP Any validates the claim against the user's "Trusted Publisher" list and hardware root-of-trust.
    5. MCP Any generates an "Attestation Token" for the MCP server.
    6. Claude Code (via MCP Any) receives the tool list along with the Attestation Tokens.
    7. Claude Code accepts the tools as "Verified" and allows execution.

## 4. Design & Architecture
* **System Flow:**
    * Discovery Provider -> Identity Validator -> Attestation Signer -> Discovery Registry.
* **APIs / Interfaces:**
    * `GET /discovery/attested`: Returns the list of discovered tools with their cryptographic attestation tokens.
    * `POST /discovery/attest`: Endpoint for local MCP servers to submit identity claims.
* **Data Storage/State:**
    * Secure store for Trusted Publisher public keys and issued Attestation Tokens.

## 5. Alternatives Considered
* **Manual User Approval (HITL):** Rejected as the primary mechanism due to friction, though it remains a fallback.
* **Cloud-Based Registry:** Rejected to maintain privacy and offline-first capabilities.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All identity claims are verified using a chain of trust leading to a local hardware enclave.
* **Observability:** Audit logs for all attestation events (success, failure, reason).

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
