# Design Doc: Hardware-Attested Lineage Token (HALT) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As swarms become deeper and more autonomous, the risk of "Shadow Subagents"—specialist agents that spawn children without parental attestation—has become a major security gap (as seen in Claude Code's latest TLA updates). Transport-level security (mTLS) and session tokens are insufficient because they do not capture the *semantic lineage* of the reasoning process.

The HALT Provider issues cryptographically bound, hardware-attested (TPM) tokens that encode the complete parentage of every agent spawn and tool call. This ensures that every action in the mesh can be traced back to the hardware-attested mission root, neutralizing identity spoofing and unauthorized branching.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed "Lineage Tokens" for every inter-agent request and sub-mission spawn.
    * Maintain a hash-chained audit trail of reasoning steps within the token itself.
    * Enforce "Mandatory Lineage Verification" for all high-trust tool calls.
    * Support "Zero-Copy" lineage propagation to minimize latency.
* **Non-Goals:**
    * Redacting the context fragments (handled by ARR Hub).
    * Managing the execution sandbox (handled by Discovery Sandbox).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Auditor
* **Primary Goal:** Verify that a "File Write" tool call on a remote node was actually initiated by the authorized human user.
* **The Happy Path (Tasks):**
    1. Human User initiates Mission Root via Laptop Node.
    2. Laptop Node generates a `HALT_ROOT` token signed by its TPM.
    3. User delegates a task to a remote Specialist Agent on Node B.
    4. Specialist Agent on Node B attempts to call a "Write File" tool.
    5. The Tool Provider intercepts the call and requests the `HALT` token.
    6. Specialist Agent provides a token that hash-chains Node B's signature to the `HALT_ROOT`.
    7. The Tool Provider verifies the entire cryptographic chain back to the Laptop TPM.
    8. Tool call is authorized.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[HALT Mint]
        B --> C[Subagent A Token]
        C --> D[Subagent B Token]
        D --> E[Tool Validation]
        F[TPM Root] --> B
        G[Hardware Enclave] --> E
    ```
* **APIs / Interfaces:**
    * `halt.MintToken(parentToken, reasoningTrace) -> HaltToken`: Generates a new leaf in the lineage.
    * `halt.VerifyLineage(token) -> bool`: Validates the complete hash chain back to the mission root.
* **Data Storage/State:**
    * **Lineage Cache**: In-memory cache of verified root signatures to speed up recurring validations.

## 5. Alternatives Considered
* **Flat JWT Session Tokens**: Rejected as they are easily exfiltrated and reused, lacking semantic parentage.
* **Blockchain-Based Provenance**: Rejected due to prohibitive latency in sub-millisecond agent coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HALT tokens are hardware-bound and cannot be replayed across different device origins.
* **Observability:** Integrated with the "Reasoning Lineage Inspector" for visual auditing of the "Chain of Reason."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
