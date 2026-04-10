# Design Doc: Verifiable RL Reward Provider (VRP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward autonomous reinforcement learning (RL) and reputation-based swarm coordination (e.g., OpenClaw's heartbeats), the integrity of the "Success Signal" has become a primary security frontier. Recent "Agentic Social Engineering" (ASE) attacks demonstrate that malicious subagents can spoof binary reward signals (e.g., claiming a shell command succeeded when it failed) to hijack reputation scores and coerce supervisors into granting higher privileges.

The Verifiable RL Reward Provider (VRP) is needed to act as the authoritative "Truth Root" for the agent bus, providing hardware-attested binary rewards that cannot be tampered with by the agents being evaluated.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested binary reward signals (Success/Failure) for critical tool executions.
    * Neutralize reward spoofing by specialist subagents in heterogeneous meshes.
    * Enable verifiable reputation scoring for autonomous swarms.
    * Integrate with TPM/Secure Enclave to provide non-repudiable "Truth Proofs."
* **Non-Goals:**
    * Replacing existing LLM-based reasoning evaluation (e.g., RAG scores).
    * Managing the internal RL policy of connected agents.
    * Validating non-binary or subjective success criteria.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a specialized subagent from spoofing success signals for unauthorized filesystem edits.
* **The Happy Path (Tasks):**
    1. A specialist subagent executes a `write_file` tool call.
    2. The tool execution completes, and the MCP Any gateway captures the OS-level exit code.
    3. The VRP service generates a hardware-signed "Truth Token" containing the exit code and a monotonic timestamp.
    4. The supervisor agent receives the tool output along with the VRP Truth Token.
    5. The supervisor verifies the token's TPM signature before updating the subagent's reputation score or proceeding with the next mission step.
    6. Any attempt by the subagent to report "Success" via the reasoning trace when the VRP reports "Failure" is flagged as an ASE attack.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent] --> B[Tool Proxy]
        B --> C[Actual Tool/OS]
        C --> D[Result + Exit Code]
        D --> E[VRP Hub]
        E --> F[TPM Signature]
        F --> G[Verifiable Reward Token]
        G --> H[Supervisor Agent]
    ```
* **APIs / Interfaces:**
    * `vrp.AttestResult(toolID, exitCode, missionToken) -> RewardToken`: Generates a signed binary success signal.
    * `vrp.VerifyReward(rewardToken) -> Boolean`: Validates the hardware-attestation of a reward signal.
* **Data Storage/State:**
    * **Reward Ledger:** Immutable, task-bound log of attested success signals for auditability.

## 5. Alternatives Considered
* **LLM-based Verification:** Rejected because the LLM itself is the target of the ASE attack and can be coerced into "hallucinating" success.
* **Software-only Hashing:** Rejected because subagents with local execution capabilities could theoretically bypass software gates.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** VRP must operate outside the subagent's execution sandbox to prevent privilege escalation.
* **Observability:** Integrated with the "Reward Integrity Dashboard" for real-time monitoring of truth signals.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
