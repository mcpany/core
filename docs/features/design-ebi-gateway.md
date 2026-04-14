# Design Doc: Enterprise-Bound Identity (EBI) Gateway
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
Recent restrictions by LLM providers (e.g., Anthropic blocking Pro/Max tokens in third-party tools as of April 4, 2026) highlight a fragility in agentic identity. When an agent's authority is tied directly to a vendor-specific subscription token, the user's agency can be revoked at any time. The EBI Gateway decouples mission-bound authority from vendor-specific tokens by acting as a framework-neutral "Identity Mint."

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) identity tokens to agents.
    * Bridge local enterprise authority with multiple LLM provider sessions.
    * Ensure non-repudiable agency that persists across vendor outages or token revocations.
    * Provide a unified governance interface for Non-Human Identities (NHI).
* **Non-Goals:**
    * Bypassing vendor Terms of Service or billing.
    * Replacing LLM-specific API keys (EBI acts as a secure wrapper/broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** CISO (Chief Information Security Officer)
* **Primary Goal:** Maintain control over agent actions in a multi-cloud environment even if an LLM provider changes its API access policy.
* **The Happy Path (Tasks):**
    1. CISO configures the EBI Gateway with local hardware-attestation requirements.
    2. An agent requests a session to perform code review.
    3. EBI Gateway issues a TPM-signed token bound to the specific mission "Code Review 2026-A".
    4. The agent presents this token to MCP Any's validation layer.
    5. MCP Any allows the agent to call local tools based on the EBI token's authority, even if the underlying Claude/GPT subscription token is restricted or rotates.
    6. All actions are logged under the persistent EBI identity, ensuring a consistent audit trail.

## 4. Design & Architecture
* **System Flow:**
    * Local Hardware Root -> EBI Gateway -> Session Token Issuance.
    * Agents carry EBI tokens in standard A2A headers.
    * MCP Any validates EBI tokens against its internal Identity Registry.
* **APIs / Interfaces:**
    * `/v1/identity/mint` (requires hardware proof)
    * `/v1/identity/verify`
* **Data Storage/State:** Encrypted local SQLite store for identity fragments and audit logs.

## 5. Alternatives Considered
* **Vendor-Specific IAM**: Rejected because it creates "Identity Silos" that are vulnerable to provider-side policy changes.
* **OpenID Connect (OIDC)**: Supported as an upstream source, but EBI adds hardware-bound attestation for "Local Sovereignty."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements "Identity-Bound Discovery"—capabilities are only revealed to verified EBI identities.
* **Observability:** Centralized audit log for all framework-neutral agent identities.

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.
