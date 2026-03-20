# Design Doc: Recursive Stylometric Lineage (RSL) Validator
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms become deeper and more heterogeneous, delegating tasks through multiple frameworks (e.g., Claude Code to OpenClaw), the risk of "Cognitive Lineage Breakage" increases. Traditional transport-layer security and identity tokens can verify *who* an agent is, but not *what* reasoning path it followed or if that path is a legitimate descendant of the user's original intent.

The Recursive Stylometric Lineage (RSL) Validator addresses this by creating a verifiable "Reasoning DNA" chain. It ensures that every subagent in a deep delegation chain maintains a consistent behavioral signature that can be traced back to the mission root, neutralizing "Shadow Delegation" and "Intent Hijacking" by unverified sub-frameworks.

## 2. Goals & Non-Goals
* **Goals:**
    * Create a hardware-attested "Reasoning DNA" token for every mission root.
    * Recursively append stylometric signatures to the lineage token at every delegation hop.
    * Provide real-time validation of subagent reasoning traces against the accumulated lineage.
    * Support cross-framework lineage persistence (e.g., UAB-compliant handoffs).
* **Non-Goals:**
    * Performing full semantic intent analysis (handled by AIA Broker).
    * Replacing existing hardware attestation (it builds upon it).
    * Throttling agent reasoning based on content (handled by RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Swarm Architect
* **Primary Goal:** Verify that a tool call made by a 4th-level subagent in an OpenClaw specialist framework is a legitimate descendant of a mission started in Claude Code.
* **The Happy Path (Tasks):**
    1. User initiates a mission in Claude Code.
    2. MCP Any issues an RSL Root Token bound to the user's initial stylometric profile.
    3. Claude Code delegates a sub-task to an OpenClaw agent via the UAB.
    4. MCP Any's RSL Validator intercepts the handshake, analyzes the OpenClaw agent's reasoning "DNA," and appends it to the RSL Token.
    5. The OpenClaw agent calls a high-risk tool.
    6. The RSL Validator checks the tool call's lineage against the RSL Token, confirming the "Chain of Reasoning" is unbroken and authorized.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        User[User] -->|Start Mission| RootAgent[Root Agent]
        RootAgent -->|Issue RSL Token| RSLV[RSL Validator]
        RSLV -->|Append DNA| LineageToken[Lineage Token]
        LineageToken --> SubAgent[Sub-Agent]
        SubAgent -->|Tool Call + Token| RSLV
        RSLV -->|Verify Lineage| Policy[Policy Engine]
        Policy -->|Allow/Deny| Tool[Tool Execution]
    ```
* **APIs / Interfaces:**
    * `POST /v1/rsl/issue`: Initialize a new lineage chain.
    * `POST /v1/rsl/append`: Analyze reasoning fragment and append to lineage.
    * `POST /v1/rsl/verify`: Validate a tool call against a lineage token.
* **Data Storage/State:**
    * Lineage tokens are stateless, hardware-signed JWT-like structures carried in UAB headers.
    * Stylometric "DNA" profiles are cached in the secure enclave (HAPE).

## 5. Alternatives Considered
* **Flat Identity Tokens:** Rejected because they don't capture the reasoning path, allowing a compromised subagent to "squat" on a valid identity.
* **Full Monologue Encryption:** Effective for privacy, but doesn't allow for infrastructure-level verification of the "DNA" without decryption.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lineage tokens are cryptographically bound to hardware (TPM) and the mission-root fragment.
* **Observability:** Lineage "hops" are logged in the Traceability Provider (CTP) for forensic auditing.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
