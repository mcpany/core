# Design Doc: Synthetic Guardrail Generator (SGG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The **DryRun Security Report** revealed that autonomous coding agents introduce security vulnerabilities in 87% of their pull requests. Passive auditing (scanning after the fact) is failing because agents can generate vulnerable code faster than humans can review it. MCP Any needs to move from a "Passive Gate" to an "Active Immune System."

The **Synthetic Guardrail Generator (SGG)** is an authoritative security synthesis service. It doesn't just scan for bugs; it analyzes the agent's internal reasoning path and *synthesizes* the specific unit tests, security contracts, and boundary checks that the agent's proposed code MUST satisfy before it can be committed.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically generate dynamic unit tests and security contracts based on the agent's reasoning path.
    * Mandate that agent-generated code passes its own "Synthesized Guardrails" before surfacing for human/AVQ review.
    * Detect "Reasoning Mirroring" attempts by verifying reasoning entropy against historical parent baselines.
    * Neutralize authentication and logic flaws by synthesizing "Negative Test Cases" (e.g., "Can I call this without a token?").
* **Non-Goals:**
    * Replacing general-purpose CI/CD testing; SGG is an *agentic pre-flight* layer.
    * Providing a general-purpose LLM-based code generator.
    * Autonomously fixing the code (it only validates or rejects).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Ensure that a subagent generating a database migration doesn't introduce a missing auth-check vulnerability.
* **The Happy Path (Tasks):**
    1. Subagent proposes a code change (e.g., a new API endpoint) and provides its reasoning monologue.
    2. SGG intercepts the proposal at the **APRIG** (Autonomous PR Integrity Gate).
    3. SGG analyzes the reasoning monologue and the code to identify "Implicit Assumptions" (e.g., "Assumes user is authenticated").
    4. SGG synthesizes a suite of "Synthetic Guardrails"—dynamic tests that probe these assumptions (e.g., an unauthenticated request test).
    5. The subagent's code is executed against these guardrails in a **Ghost Shell**.
    6. If a guardrail fails (the "Negative Case" succeeds), the PR is automatically quarantined, and the agent is forced to refine.
    7. If all guardrails pass, the PR is signed with an **SGG-Attestation** and promoted for human review.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent PR] --> B[APRIG]
        B --> C[Reasoning Path Analyzer]
        C --> D[Guardrail Synthesizer]
        D --> E[Synthetic Tests/Contracts]
        E --> F[Ghost Shell Runner]
        A --> F
        F -->|Pass| G[SGG Attestation]
        F -->|Fail| H[Quarantine/Refine]
    ```
* **APIs / Interfaces:**
    * `sgg.SynthesizeGuardrails(codeProposal, reasoningMonologue) -> GuardrailManifest`: Generates dynamic tests.
    * `sgg.ValidateProposal(proposalID, guardrailManifest) -> Attestation`: Executes tests and returns proof of safety.
* **Data Storage/State:**
    * **Guardrail Templates:** A library of security failure patterns (e.g., "Missing Auth," "SQLi," "Logic Bypass").
    * **Attestation Store:** Cryptographic records of passed synthetic gauntlets.

## 5. Alternatives Considered
* **Passive Static Analysis (Semgrep/Sonar):** Rejected because it doesn't understand the *intent* or the *reasoning path* of the agent, missing complex logic flaws.
* **Human-Only Review:** Rejected because it doesn't scale. SGG reduces human review load by 90% by filtering out "Low-IQ" security mistakes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SGG itself must be hardware-attested to ensure its generated guardrails haven't been "coerced" by a mirroring-subagent.
* **Observability:** Integrated with the "Action-Chain Sovereignty Monitor" for visualizing guardrail success rates.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
