# Design Doc: Attested Discovery Authority
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
With the rise of "Configuration-as-Execution" exploits and mass remote exposure of agent gateways (OpenClaw), tool discovery has become a primary attack vector. MCP Any needs to act as a "Certificate Authority" for tools, providing cryptographic proof of a tool's provenance before it is ingested by the agent runtime.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a mechanism for MCP servers to sign their capability cards.
    * Enable agent frameworks (e.g., Claude Code) to verify tool identity before ingestion.
    * Neutralize "Shadow Capability" mapping by unauthenticated tools.
* **Non-Goals:**
    * Replacing the MCP protocol itself.
    * Managing tool-specific credentials (API keys).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure that only pre-approved, signed tools can be discovered and executed by internal agent swarms.
* **The Happy Path (Tasks):**
    1. MCP Server generates a CSR (Certificate Signing Request) for its capability card.
    2. MCP Any Attestation Authority signs the card using a hardware-bound key (TPM).
    3. Agent requests tool discovery from MCP Any.
    4. MCP Any serves only the signed and verified capability cards.
    5. Agent framework verifies the signature against MCP Any's public root.

## 4. Design & Architecture
* **System Flow:**
    [MCP Server] --(Signed Card)--> [MCP Any Discovery Hub] --(Verified List)--> [Agent Runtime]
* **APIs / Interfaces:**
    * `POST /v1/attest/tool`: Submit tool metadata for signing.
    * `GET /v1/discovery/signed`: Retrieve the list of attested tools.
* **Data Storage/State:**
    * Attested tools are stored in the Shared KV Store (Blackboard) with a `trusted: true` label.

## 5. Alternatives Considered
* **Static Allow-lists**: Rejected due to the dynamic nature of agent swarms and marketplace plugins.
* **Manual HITL Approval**: Too slow for autonomous swarm coordination; manual approval is moved to the "Sign-off" phase rather than the "Discovery" phase.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All discovery requests mandate A2A handshakes.
* **Observability:** Audit logs for every signature event and discovery request.

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
