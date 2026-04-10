# Design Doc: Agentic Quorum Hub (AQC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents become more autonomous, Human-in-the-Loop (HITL) becomes a bottleneck for high-frequency, high-stakes operations. The Agentic Quorum Hub (AQC) evolves HITL into "Agent-in-the-Loop," where a cryptographically bound quorum of specialized auditor agents must approve a tool call before execution. This ensures collective cognitive sovereignty and prevents a single compromised specialist from performing unauthorized actions.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a distributed consensus middleware for tool-call validation.
    * Support "Sovereignty Impact" scoring to automatically trigger quorum requirements.
    * Provide hardware-attested approval tokens for every quorum member.
    * Neutralize "Agentic Social Engineering" via multi-agent cross-verification.
* **Non-Goals:**
    * Replacing manual HITL for critical user-defined safety gates.
    * Managing the underlying LLM reasoning for auditor agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Authorize a high-risk shell command without manual human intervention while maintaining Zero-Trust.
* **The Happy Path (Tasks):**
    1. Specialist Agent requests `run_shell_command` with a high-impact payload.
    2. AQC Middleware intercepts the request and calculates a Sovereignty Impact score.
    3. Score exceeds threshold; AQC initiates a quorum request to three independent Auditor Agents.
    4. Auditor Agents receive the intent, payload, and reasoning trace.
    5. Each Auditor Agent performs independent validation and provides a hardware-signed approval token.
    6. AQC Hub aggregates tokens and verifies the cryptographic signatures.
    7. Quorum reached; the tool call is authorized and executed.

## 4. Design & Architecture
* **System Flow:**
    `Tool Request` -> `Impact Scorer` -> `Quorum Orchestrator` -> `Auditor Verification` -> `Consensus Aggregator` -> `Authorized Execution`
* **APIs / Interfaces:**
    * `QuorumHub`: `RequestQuorum(intent string, payload bytes) (QuorumID, error)`
    * `AuditorNode`: `Validate(request QuorumRequest) (ApprovalToken, error)`
* **Data Storage/State:**
    * Quorum states and signed tokens are stored in a session-bound, encrypted buffer.

## 5. Alternatives Considered
* **Single Auditor Agent**: Rejected because it creates a single point of failure (if the auditor is compromised).
* **Deterministic Policy Engine (Rego)**: Useful for static rules, but rejected as a standalone solution because it cannot evaluate the "Semantic Intent" of complex reasoning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All quorum communication is encrypted and hardware-attested.
* **Observability:** Quorum decisions, member votes, and impact scores are logged for forensic auditing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
