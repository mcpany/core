# Design Doc: Zero-Knowledge State Attestation (ZKSA) Provider
**Status:** Draft
**Created:** 2026-07-10

## 1. Context and Scope
As agent swarms move from single-tenant to multi-tenant environments, the need for "Verifiable Privacy" has become critical. Specialist agents often process sensitive data (PII, trade secrets) that should not be exposed to the parent agent or the central coordination bus. However, the system still needs to verify that these agents are operating within safety constraints and conforming to mission-root policies.

The ZKSA Provider enables subagents to generate cryptographic proofs that their internal state aligns with a verified mission manifest or security schema without revealing the raw data itself. This allows for Zero-Trust coordination where agents can verify each other's integrity without compromising data sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a proof-broker for Zero-Knowledge State Attestation.
    * Provide standardized ZK-circuits for common state-conformance checks (e.g., "Contains no PII," "Matches schema X").
    * Integrate with the SRM Provider to anchor proofs to hardware-attested sessions.
    * Enable sub-millisecond proof verification at the coordination gateway.
* **Non-Goals:**
    * Implementing custom LLM reasoning logic.
    * Providing general-purpose ZK computation (focus is strictly on agent state attestation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Data Privacy Auditor
* **Primary Goal:** Verify that a "Code Reviewer" agent has not ingested sensitive database credentials without actually seeing the agent's internal reasoning monologue.
* **The Happy Path (Tasks):**
    1. The parent agent spawns a "Code Reviewer" subagent with a ZK-Policy requiring "Credential-Free Context."
    2. The subagent processes the code and generates a ZK-proof of policy compliance using the ZKSA Provider.
    3. The subagent sends the proof to the MCP Any gateway along with its task completion signal.
    4. MCP Any verifies the proof against the hardware-bound mission-root manifest.
    5. The gateway authorizes the task merge, cryptographically certain that no credentials were leaked, despite never seeing the subagent's raw memory.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Subagent[Specialist Subagent] -->|Generate Proof| ZKSA[ZKSA Provider]
        ZKSA -->|Signed Proof| Gateway[MCP Any Gateway]
        Gateway -->|Verify| Policy[Mission Root Policy]
        Policy -->|Authorized| Blackboard[Blackboard Commit]
        Subagent -.->|Private State| LocalMem[Local Secure Enclave]
    ```
* **APIs / Interfaces:**
    * `POST /zksa/proof/generate`: Circuit-bound proof generation for subagents.
    * `POST /zksa/proof/verify`: High-speed verification endpoint for the gateway.
    * `GET /zksa/circuits`: Registry of available state-conformance circuits.
* **Data Storage/State:**
    * Public verification keys are stored in the Hardware-Attested Mission Manifest.
    * Private witness data never leaves the subagent's isolated execution environment.

## 5. Alternatives Considered
* **Plaintext Scrubbing:** Rejected due to the risk of "Instruction Smuggling" and the inability to provide cryptographic guarantees.
* **Fully Homomorphic Encryption (FHE):** Rejected due to current performance overhead (100x+ latency) which is unacceptable for real-time agent coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZK-proofs must be hardware-bound to the TPM session to prevent proof-replay attacks.
* **Observability:** Verification successes and failures are logged in the Mission Audit Trail with high-resolution timestamps.

## 7. Evolutionary Changelog
* **2026-07-10:** Initial Document Creation.
