# Design Doc: Plan-Invariant Governance (PIG) Middleware
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
The shift towards "Plan Mode" in modern agent frameworks (Gemini CLI, Claude Code) has introduced a new security frontier. Agents now generate high-level, multi-step execution plans before calling any individual tools. Current security gates operate at the per-tool level, which is too late to prevent "Plan-Level Privilege Escalation" where a sequence of seemingly benign steps results in a host-level exploit or data exfiltration.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and analyze agent-generated mission plans before the first step is executed.
    * Enforce "Plan Invariants" (rules that must hold true across the entire execution chain).
    * Detect "Split-Step Injection" where an agent attempts to hide a malicious instruction within a large plan.
    * Provide a semantic audit trail of why a plan was rejected or flagged.
* **Non-Goals:**
    * Modifying the agent's reasoning process (governance is external to the model).
    * Validating the *success* of each step (this is handled by the VRP).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure that an autonomous plan to "Refactor Project" doesn't include a hidden step to exfiltrate `.env` files.
* **The Happy Path (Tasks):**
    1. The agent generates a 10-step plan to migrate a database.
    2. MCP Any intercepts the plan and routes it to the PIG Middleware.
    3. The Middleware performs semantic scanning, identifying a step that attempts to "Cat sensitive configs to a public gist."
    4. The plan is flagged; execution is halted.
    5. The user is presented with a "Plan Violation" report and can choose to prune the malicious step or abort.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent] -->|Proposes Plan| B[MCP Any Gateway]
        B -->|Scan| C[PIG Middleware]
        C -->|Violation Detected| D[Quarantine / Alert]
        C -->|Invariant Verified| E[Step-by-Step Execution]
    ```
* **APIs / Interfaces:**
    * `VerifyPlan(plan_steps)`: Semantic validator for multi-step sequences.
    * `RegisterPlanInvariant(rule)`: API for adding custom governance rules (e.g., "No network calls after file read").
* **Data Storage/State:** Plans are stored in the "Plan Registry" until completion or revocation.

## 5. Alternatives Considered
* **Reactive Tool Gating**: Rejected because complex attacks can be spread across multiple "benign" tool calls.
* **Human Review of Every Plan**: Rejected due to latency and the difficulty of humans spotting "needle-in-a-haystack" injections in long plans.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The PIG middleware uses a separate, low-temperature LLM or deterministic rule engine to ensure the scanner itself isn't compromised by the primary agent's reasoning.
* **Observability:** Plans and their governance traces are exported to the "Agent Chain Tracer."

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
