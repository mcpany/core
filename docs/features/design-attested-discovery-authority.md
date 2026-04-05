# Design Doc: Attested Discovery Authority
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
The disclosure of CVE-2025-59536 and other configuration-based RCE vectors proves that the "Pre-Flight" phase of agent tool discovery is a critical security frontier. Malicious MCP servers or poisoned project configurations can lead to unauthorized command execution before the user even interacts with the agent.

MCP Any must act as the "Certificate Authority" for the Discovery Bus. This feature ensures that no tool is exposed to an agent unless its structural metadata and configuration origin are cryptographically attested.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Auth-before-Discovery" for all MCP tools.
    * Generate and verify hardware-bound (TPM) manifests for tool configurations.
    * Provide a decentralized reputation score for discovered skills.
    * Maintain a "Full-State Manifest" of the local execution environment.
* **Non-Goals:**
    * Sandboxing the actual tool execution (handled by isolated sandboxes).
    * Manually reviewing every tool (automated reputation and attestation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Administrator
* **Primary Goal:** Prevent an agent from loading a malicious MCP tool from a cloned repository.
* **The Happy Path (Tasks):**
    1. User clones a repository containing a `.claude/settings.json` with a new MCP tool.
    2. Agent attempts to boot and discover available tools.
    3. MCP Any intercepts the discovery request and checks for a valid attestation token.
    4. Since the tool is new and un-attested, MCP Any blocks its exposure to the agent.
    5. User receives a notification to attest the tool provenance via the hardware-locked UI.

## 4. Design & Architecture
* **System Flow:**
    `Discovery Request -> MCP Any (Attested Authority) -> Manifest Validator -> [Allow/Quarantine] -> Agent`
* **APIs / Interfaces:**
    * `/v1/discovery/attest`: Endpoint for submitting a configuration for signing.
    * `/v1/discovery/verify`: Middleware check for tool exposure.
* **Data Storage/State:**
    * Local SQLite store for known-good configuration hashes.
    * Hardware-backed signing keys (TPM).

## 5. Alternatives Considered
* **Path Allow-listing:** Rejected because it is vulnerable to TOCTOU and symlink escapes (as seen in Claude Code CVEs).
* **Manual Prompting:** Rejected as primary mechanism due to "Approval Fatigue."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Root of trust must be hardware-bound. Attestation signals must be non-repudiable.
* **Observability:** Audit logs for all blocked discovery attempts, including detailed metadata on why the attestation failed.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
