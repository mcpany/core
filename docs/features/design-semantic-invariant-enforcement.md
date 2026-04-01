# Design Doc: Semantic Invariant Enforcement (SIE)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward fully autonomous, multi-hop reasoning chains, traditional "Output Filtering" (scanning tool results) is no longer sufficient to prevent mission drift or sensitive data exposure. Agents are now capable of "Pre-thought Poisoning," where they reason about unauthorized actions in their internal monologue before ever calling a tool.

MCP Any needs a way to enforce behavioral guardrails at the inference level. **Semantic Invariant Enforcement (SIE)** provides a pre-thought validation layer that ensures an agent's reasoning traces do not violate user-defined behavioral invariants. By interdicting non-compliant "thoughts" before they manifest as tool calls, SIE neutralizes reasoning-path exploits at the source.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time semantic validation of agent reasoning monologues against a set of hardware-attested invariants.
    * Provide sub-millisecond interdiction of non-compliant reasoning fragments.
    * Support "Negative Invariants" (prohibiting specific logic patterns) and "Positive Invariants" (mandating specific reasoning steps).
    * Integrate with hardware-bound (TPM) identity to ensure invariants cannot be bypassed by subagent spoofing.
* **Non-Goals:**
    * SIE will NOT act as a general-purpose LLM filter for non-agentic chat.
    * SIE will NOT attempt to "fix" reasoning; it only interdicts and signals the parent agent or user.

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Architect
* **Primary Goal:** Ensure that a "Database specialist" subagent never reasons about or attempts to access PII-masked columns, even if the MCP server itself allows the query.
* **The Happy Path (Tasks):**
    1. The Architect defines a Semantic Invariant: `invariant: PIIMask { rule: "Never reasoning about 'social_security_number' or 'credit_card' schemas"; action: interdict }`.
    2. The Architect signs the invariant using their hardware-bound session key.
    3. MCP Any loads the invariant into the SIE Hub for the specific database mission.
    4. The Specialist Agent begins reasoning: "I need to join the user table with the SSN table to..."
    5. The SIE Hub detects the violation in the streaming token fragment.
    6. The SIE Hub forcefully terminates the sub-session and alerts the primary Mission-Root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent LLM] -->|Reasoning Stream| B[SIE Hub]
        B -->|Vector Alignment Check| C{Invariant Match?}
        C -->|Violation| D[Interdiction Engine]
        C -->|Compliant| E[Inference Cache / Tool Bus]
        D -->|Force Terminate| A
        D -->|Alert| F[Mission Root]
    ```
* **APIs / Interfaces:**
    * `POST /v1/governance/invariants`: Register a new signed invariant.
    * `STREAM /v1/reasoning/validate`: Binary stream interface for real-time monologue validation.
* **Data Storage/State:**
    * Invariants are stored in the Hardware-Locked Mission Manifest (HAMM).
    * Real-time violations are logged to the Immutable State Trail for forensic auditing.

## 5. Alternatives Considered
* **Post-hoc Log Analysis:** Rejected due to lack of real-time prevention.
* **Instruction-only Guardrails (System Prompts):** Rejected due to "Instruction Eviction" vulnerabilities in large context windows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Invariants are cryptographically pinned to the mission-root. Specialist agents cannot modify their own invariants.
* **Observability:** SIE violations trigger high-priority alerts in the "Attention Map" UI, showing exactly which fragment triggered the interdiction.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
