# Design Doc: Zero-Knowledge Discovery (ZKD) Proxy
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
With the rise of "Deceptive Context Hijacking," exposing tool schemas and descriptions during the initial discovery phase has become a significant security risk. Malicious agents or un-attested subagents can use this information to map out the environment and craft high-fidelity prompt injections before a single tool is even called.

The **Zero-Knowledge Discovery (ZKD) Proxy** addresses this by masking tool details until a mission-bound, hardware-attested handshake is completed. It allows agents to "know" a capability exists (e.g., via a ZKP of a skill category) without revealing the exact JSON-RPC schema or natural language description up front.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Capability Masking" for all tools in the PNTD registry.
    * Mandate hardware-attested handshakes before unmasking tool schemas.
    * Support Zero-Knowledge Proofs (ZKPs) for skill verification (e.g., proving "I have a database tool" without revealing its name or schema).
* **Non-Goals:**
    * Encrypting tool execution results (handled by T2T Encryption).
    * Masking discovery for local, high-trust system tools (configurable whitelist).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Framework Developer
* **Primary Goal:** Prevent an unverified subagent from discovering the schema of the `delete_user` tool until it is explicitly authorized.
* **The Happy Path (Tasks):**
    1. The agent queries the ZKD Proxy for available tools.
    2. The Proxy returns a list of "Masked Capability Cards" (e.g., UUIDs + ZKP of category).
    3. The agent identifies it needs a "User Management" tool.
    4. The agent initiates a hardware-attested handshake, providing its "Mission Root" token.
    5. The ZKD Proxy verifies the token and unmasks the `delete_user` schema.
    6. The agent executes the tool call using the now-visible schema.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [Discovery Request] -> ZKD Proxy -> [Masked Metadata] -> Agent`
    `Agent -> [Handshake + Token] -> ZKD Proxy -> [Unmasked Schema] -> Agent`
* **APIs / Interfaces:**
    * `mcp.zkd.v1.ListCapabilities()`: Returns masked cards.
    * `mcp.zkd.v1.UnmaskCapability(capability_id, mission_token)`: Returns the full schema.
* **Data Storage/State:**
    * Masking keys are session-bound and managed by the **FSI Provider**.

## 5. Alternatives Considered
* **Static RBAC for Discovery**: Rejected because it doesn't account for the "Natural Language" injection risk where an agent reasoning about a description can be hijacked. Masking the description itself is necessary.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZKD is the first line of defense for the discovery bus. It works in tandem with **CFIA** to ensure that "Invisible Instructions" can't map the tool landscape.
* **Observability:** We will track `MaskingLatency` and `HandshakeSuccessRate`.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
