# Design Doc: Per-Agent Identity Mint (PAIM)
**Status:** Draft
**Created:** 2026-04-15

## 1. Context and Scope
Agent frameworks currently rely on shared, unscoped API keys for tool access. As swarms become deeper and more autonomous, the lack of unique agent identities makes it impossible to enforce granular permissions or perform non-repudiable auditing. PAIM solves this by providing a hardware-attested identity layer that wraps existing frameworks.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue unique, task-bound identity tokens for every subagent.
    * Provide hardware-attested lineage for all issued tokens.
    * Support retroactive identity injection for existing frameworks (OpenClaw, AutoGen).
* **Non-Goals:**
    * Replacing existing LLM provider authentication (API keys). PAIM acts as a scope *on top* of existing keys.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Swarm Architect
* **Primary Goal:** Audit exactly which specialist agent in a 5-agent mesh initiated a sensitive database write.
* **The Happy Path (Tasks):**
    1. The parent agent requests a subagent spawn via MCP Any.
    2. PAIM generates a hardware-attested "Identity Card" for the subagent, bound to the parent's mission.
    3. The subagent performs a tool call, including its PAIM token in the request header.
    4. MCP Any validates the token and logs the action against the unique subagent ID.

## 4. Design & Architecture
* **System Flow:**
    `Agent Framework` -> `PAIM Token Request` -> `TPM/Secure Enclave Signature` -> `PAIM Token Issued` -> `Tool Call with Token` -> `MCP Any Validation`
* **APIs / Interfaces:**
    * `/v1/identity/mint`: Issues a new task-bound token.
    * `/v1/identity/verify`: Validates a token and returns the agent lineage.
* **Data Storage/State:**
    Tokens are stored in the secure Blackboard with hardware-bound expiry.

## 5. Alternatives Considered
* **Framework-native identity:** Rejected because most frameworks (93%) currently lack any scoped identity, and MCP Any aims to be a universal adapter.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are short-lived and cryptographically bound to the hardware root of trust.
* **Observability:** Every tool call is now linked to a specific PAIM identity in the audit logs.

## 7. Evolutionary Changelog
* **2026-04-15:** Initial Document Creation.
