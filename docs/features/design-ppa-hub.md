# Design Doc: Privacy-Preserving Audit (PPA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents become more autonomous, the need for security auditing increases. However, traditional logs often contain sensitive "Internal Monologues" or mission-critical context that should not be exposed to external auditors or even centralized logging systems. Gemini CLI's introduction of Privacy-Preserving Reason Proofs (PPRP) highlights the need for a system that proves an agent's reasoning integrity without revealing its raw context.

The PPA Hub acts as a Zero-Knowledge proof broker, allowing agents to attest that their high-risk actions (and the reasoning behind them) are compliant with the mission root without exfiltrating sensitive data.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate the generation and verification of Zero-Knowledge "Reasoning Proofs" for high-risk tool calls.
    * Enable **Audit-Gated Discovery (AGD)**, where tool schemas are only revealed post-verification of reasoning integrity.
    * Provide a verifiable audit trail that satisfies SSDF and SSCC compliance without raw data exposure.
    * Neutralize "Reasoning Drift" by enforcing cryptographic alignment between the reasoning proof and the final tool input.
* **Non-Goals:**
    * Storing raw reasoning traces; the Hub only stores the cryptographic proofs and attestation results.
    * Managing the model's generation process; it acts as a post-generation / pre-execution gate.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Auditor
* **Primary Goal:** Verify that a swarm of 50 agents performing automated system patches did not attempt any unauthorized file exfiltration, without reading the agents' private logs.
* **The Happy Path (Tasks):**
    1. Agent generates a reasoning path for a high-risk patch operation.
    2. Agent invokes the PPA Hub to generate a Reasoning Proof (PPRP).
    3. PPA Hub utilizes a local hardware enclave to verify the reasoning against the Mission Root manifest.
    4. Hub issues a signed "Reasoning Attestation" token.
    5. Agent presents this token to the MCP Any gateway to "Unlock" the capability to call the `patch_system` tool.
    6. MCP Any executes the tool and logs the Attestation Token ID.
    7. External Auditor queries the Hub using the Token ID to see the "Compliant" status and hardware signature, without ever seeing the raw Reasoning Monologue.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Monologue] --> B[PPA Hub / Local Enclave]
        B -->|ZKP Generation| C[Reasoning Proof]
        C --> D[Audit-Gated Discovery Bus]
        D -->|Reveal Tool| E[Agent Execution]
        E --> F[Audit Ledger]
    ```
* **APIs / Interfaces:**
    * `ppa.GenerateProof(monologue, missionRoot) -> ProofID`: Generates a hardware-attested ZK proof.
    * `ppa.VerifyProof(proofID) -> AttestationToken`: Verifies the proof and issues an execution capability.
    * `ppa.QueryAudit(tokenID) -> ComplianceReport`: Returns a privacy-safe audit signal.
* **Data Storage/State:**
    * **Proof Registry:** Hardware-locked database of proof IDs and their compliance metadata.
    * **Audit Ledger:** Append-only log of tool calls linked to their corresponding Attestation Tokens.

## 5. Alternatives Considered
* **Manual Log Review:** Rejected due to lack of scalability and extreme privacy risks.
* **Sentiment/Safety Model Scanning:** Rejected because it is non-deterministic and can be bypassed by "Reasoning toward the Audit" (stylistic mimicry). PPA Hub requires cryptographic alignment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Hub itself runs in a hardware-isolated environment to prevent proof tampering.
* **Observability:** Integrated with the "Zero-Knowledge Audit Viewer" UI for real-time compliance monitoring.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
