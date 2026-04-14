# Design Doc: Hardware-Bound Per-Agent Identity (HPAI)
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
Current AI agent frameworks (93%) rely on unscoped API keys, and 0% provide unique identities for individual agents. This lack of identity leads to "Agentic Social Engineering" and unauthorized resource access where a compromised subagent can act with the full authority of its parent without any traceable lineage.

MCP Any must solve this by acting as a hardware-attested identity mint, ensuring every agent in the mesh has a unique, non-repudiable identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-backed identity tokens to every subagent instance.
    * Provide a non-repudiable audit trail of tool calls per agent.
    * Enable framework-agnostic identity persistence across handoffs.
* **Non-Goals:**
    * Replacing existing LLM provider authentication (e.g., Anthropic API keys).
    * Providing user-level identity management (this is per-agent identity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Audit a multi-agent workflow to identify which specific subagent triggered a high-risk tool call.
* **The Happy Path (Tasks):**
    1. Architect defines a "Mission Root" in MCP Any.
    2. Primary Agent spawns 3 specialist subagents.
    3. MCP Any's HPAI provider issues unique, TPM-signed identity tokens to each subagent.
    4. Specialist Agent A calls `run_shell_command`.
    5. MCP Any validates the token and logs the call against Specialist Agent A's unique ID.
    6. Architect views the audit log and sees the call attributed specifically to Specialist Agent A, not the generic parent key.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Spawn] -> [HPAI Token Request] -> [TPM Attestation] -> [Identity Minting] -> [Token Injection]
* **APIs / Interfaces:**
    * `POST /v1/identity/mint`: Request a new hardware-bound identity.
    * `GET /v1/identity/verify`: Verify a token's hardware provenance.
* **Data Storage/State:**
    * Identities are stored in a hardware-protected enclave (TPM).
    * Mappings between Mission Root and Agent IDs are stored in the secure Blackboard.

## 5. Alternatives Considered
* **Software-only JWTs:** Rejected because they can be easily cloned or leaked by compromised subagents.
* **IP-based Identity:** Rejected because multiple agents often run in the same local execution environment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are short-lived and mission-bound.
* **Observability:** Every tool call includes the HPAI-ID in the structured logs.

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
