# Design Doc: Attested Discovery Authority
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
With the release of Claude Code's hardened security model (CVE-2025-59536), agentic tools now require a "Trust Signal" for every MCP server they connect to. Existing auto-discovery mechanisms are "Promiscuous by Default," allowing any local process to register as a tool. MCP Any needs to provide an `Attested Discovery Authority` that acts as the cryptographic root of trust for all discovered MCP servers, satisfying the verification requirements of advanced agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographic identity for every registered MCP server.
    * Integrate with the local TPM or Secure Enclave to sign discovery manifests.
    * Implement a "Trust Verification" endpoint that agents (like Claude Code) can query before ingesting tools.
    * Support "Zero-Trust" local discovery where tools are quarantined until a valid attestation is provided.
* **Non-Goals:**
    * Replacing the MCP protocol itself.
    * Managing remote cloud-based tool registries (this is for local/federated discovery).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer using Claude Code.
* **Primary Goal:** Connect a local database tool to Claude Code without triggering "Unverified Tool" warnings or security blocks.
* **The Happy Path (Tasks):**
    1. User starts a local MCP server.
    2. MCP Any detects the server and challenges it for identity (e.g., a signed binary hash).
    3. The `Attested Discovery Authority` verifies the signature against a local allow-list.
    4. MCP Any signs a "Discovery Attestation Token" for the server.
    5. Claude Code queries MCP Any for available tools.
    6. MCP Any provides the tool list along with the Attestation Tokens.
    7. Claude Code verifies the tokens and allows the tools to be used immediately.

## 4. Design & Architecture
* **System Flow:**
    `[Local MCP Server] --(Identity Claim)--> [Discovery Authority] --(Signed Token)--> [MCP Any Registry] --(Attested Manifest)--> [AI Agent]`
* **APIs / Interfaces:**
    * `GET /v1/discovery/attest/:server_id`: Retrieve the attestation token for a specific server.
    * `POST /v1/discovery/challenge`: Submit a new identity claim for verification.
* **Data Storage/State:**
    * `trust_registry.db`: Stores signed hashes and provenance data for verified servers.

## 5. Alternatives Considered
* **User-Manual Whitelisting**: Rejected due to high friction and the inability of agents to programmatically verify the manual entry.
* **Centralized Cloud Registry**: Rejected to maintain "Local-First" privacy and ensure the system works in air-gapped environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All tool discovery is "Block-by-Default." Un-attested tools are hidden from the agent until explicitly approved.
* **Observability**: The UI will show a "Trust Pedigree" for every tool, including the signature chain and last verification timestamp.

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
