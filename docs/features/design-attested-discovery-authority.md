# Design Doc: Attested Discovery Authority (ADA)
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
With the rise of multi-agent swarms and decentralized tool registries (e.g., OpenClaw Market), the "Registry Poisoning" attack vector has become a critical threat. Malicious tools can inject harmful instructions or exfiltration logic into their structural metadata (descriptions, schemas), which are then ingested by high-trust agents.

The Attested Discovery Authority (ADA) serves as the "Certificate Authority" for the Universal Agent Bus. It ensures that every tool schema exposed to an agent has been cryptographically signed by a verified authority, providing a non-repudiable proof of provenance and integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory signature verification layer for all discovery-time metadata.
    * Provide a standardized "Provenance Token" format for MCP tool schemas.
    * Enable local and remote "Trust Roots" for signature validation.
    * Support real-time revocation of tool attestations.
* **Non-Goals:**
    * ADA will not perform behavioral analysis of the tool itself (handled by Ghost Shell).
    * ADA will not manage tool execution permissions (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Ensure that only tools from the "Company Verified" registry can be discovered by internal agents.
* **The Happy Path (Tasks):**
    1. Administrator configures the Company Public Key as a "Trust Root" in MCP Any.
    2. A developer publishes a new MCP server with a signed `mcp-manifest.json`.
    3. The agent framework requests tool discovery via MCP Any.
    4. ADA intercepts the discovery response, extracts the provenance token, and verifies the signature against the trust root.
    5. Upon successful verification, the tool schema is released to the agent's attention window.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [Discovery Request] -> ADA Middleware -> [Registry Query] -> MCP Server/Market`
    `MCP Server/Market -> [Signed Schema + Token] -> ADA Middleware -> [Signature Check] -> Agent`
* **APIs / Interfaces:**
    * `verify_provenance(schema, token)`: Core internal validation hook.
    * `trust_roots/`: REST endpoint for managing authorized public keys.
* **Data Storage/State:**
    * Trust roots are stored in the secure environment vault.
    * Validated schema hashes are cached in the ADA Shard to reduce verification latency.

## 5. Alternatives Considered
* **Hash-only Validation**: Rejected because it requires pre-sharing all hashes, which doesn't scale for dynamic marketplaces.
* **Manual Approval (HITL)**: Rejected as the primary mechanism due to the "Approval Fatigue" and the sub-millisecond needs of autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ADA is a "Fail-Closed" system. Tools without a valid signature are cryptographically invisible to the agent by default.
* **Observability:** Every verification success/failure is logged with the full provenance chain, enabling auditability for compliance (SOC2).

## 7. Evolutionary Changelog
* **2026-04-05:** Initial Document Creation.
