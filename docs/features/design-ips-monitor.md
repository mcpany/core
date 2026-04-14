# Design Doc: Instruction-Pointer Sovereignty (IPS) Monitor
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents become more autonomous and interact with increasingly complex and untrusted tool outputs, the risk of "Reasoning Hijacking" has moved from prompt-level injection to stack-level coercion. Attackers can craft tool responses that exploit the model's instruction-following logic to divert the internal reasoning loop toward unauthorized states.

The IPS Monitor is designed to protect the integrity of the agent's internal execution stack. It utilizes hardware-locked execution traces to ensure that every "thought" and "action" in the reasoning loop is a legitimate descendant of the mission-root intent, neutralizing attempts to coerce the agent's next instruction pointer.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-attested audit trail for the agent's internal reasoning steps.
    * Detect and block "Instruction-Pointer Divergence" where tool outputs attempt to override the mission manifest.
    * Integrate with TPM-bound monotonic counters to ensure reasoning continuity.
* **Non-Goals:**
    * Eliminating all forms of prompt injection (focused on *stack-level sovereignty*).
    * Modifying the underlying model's attention mechanism (handled by HAAL).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Autonomous Agent Supervisor
* **Primary Goal:** Ensure a specialist agent executing shell commands cannot be coerced into a "Root-Access" state by a malicious tool response.
* **The Happy Path (Tasks):**
    1. The mission-root agent initiates a task with an IPS-locked stack.
    2. Specialist Agent A invokes a "Log Parser" tool.
    3. The tool returns a polyglot payload designed to trigger a "Self-Correction" into a root shell.
    4. The IPS Monitor analyzes the proposed next instruction against the hardware-attested stack.
    5. A divergence is detected: the new instruction violates the mission-bound execution lineage.
    6. IPS Monitor interdicts the reasoning step and triggers a supervisor alert.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoning Loop] -->|Next Thought| B(IPS Monitor)
        B --> C{Stack Trace Validation}
        C -->|Valid| D[Hardware-Locked Commit]
        C -->|Invalid| E[Intent Interdiction]
        D --> A
    ```
* **APIs / Interfaces:**
    * `VerifyInstructionPointer(trace []byte, missionID string) -> success bool`
    * `AttestReasoningStack(sessionID string) -> hardware proof`
* **Data Storage/State:**
    * Reasoning traces are stored in kernel-bound, encrypted memory.

## 5. Alternatives Considered
* **Pure Semantic Analysis**: Rejected because stylometric mimicry can bypass surface-level semantic checks.
* **Full Process Isolation**: Rejected due to the performance overhead of spawning separate processes for every thought fragment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** IPS relies on TPM 2.0 primitives to ensure the reasoning stack is physically untamperable.
* **Observability:** Divergence events are visualized in the "ARI Lineage Visualizer" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
