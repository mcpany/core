# Design Doc: Advisor-Executor Coordination (AEC) Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agentic workflows scale to long-horizon tasks, using high-intelligence (and often high-latency/high-cost) models for every step becomes inefficient. The "Advisor-Executor" pattern (e.g., Anthropic's Advisor Tool) addresses this by pairing a fast "Executor" model with a strategic "Advisor" model. MCP Any needs to provide the infrastructure to coordinate these pairs, ensuring that strategic guidance is seamlessly injected into the execution loop while maintaining mission-root sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate synchronization between fast executor models and high-intelligence advisor models.
    * Provide "Advisor-in-the-Loop" checkpoints for mid-generation strategic guidance.
    * Enforce mission-bound constraints on advisor suggestions to prevent reasoning drift.
    * Optimize token usage by gating advisor invocations based on task complexity or entropy.
* **Non-Goals:**
    * Directly managing LLM API keys (handled by existing transport layers).
    * Defining the internal reasoning logic of the models.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agentic Systems Architect
* **Primary Goal:** Efficiently execute a complex multi-file refactor using a fast model while relying on a strategic model for architectural guidance.
* **The Happy Path (Tasks):**
    1. Executor agent begins the refactoring task.
    2. AEC Middleware monitors the generation and identifies a "Strategic Pivot Point" (e.g., a high-entropy decision).
    3. The Middleware suspends the Executor and queries the Advisor model with the current context and strategic prompt.
    4. The Advisor returns a guidance fragment (e.g., "Use the Factory pattern here").
    5. AEC Middleware validates the guidance against the Mission Manifest and injects it into the Executor's context.
    6. The Executor resumes generation, now grounded in strategic guidance.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[AEC Middleware]
        B --> C[Executor Model]
        B --> D[Advisor Model]
        C -->|State/Entropy| B
        B -->|Guidance| C
        D -->|Strategy| B
    ```
* **APIs / Interfaces:**
    * `aec.RegisterPair(executorID, advisorID, policy) -> SessionID`: Configures a coordination session.
    * `aec.RequestGuidance(sessionID, context) -> Guidance`: Explicitly triggers advisor feedback.
    * `aec.Intercept(executorStream) -> ModifiedStream`: Automatically injects guidance into active streams.
* **Data Storage/State:**
    * **Coordination Session Store:** Tracking state and guidance history for active pairs.

## 5. Alternatives Considered
* **Manual Agent Prompting:** Rejected because it lacks standardized infrastructure and increases latency due to un-optimized context handling.
* **Linear Model Chaining:** Rejected as it doesn't allow for real-time, "mid-generation" strategic interdiction.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Advisor suggestions are treated as untrusted input and must pass through the "Injection-Shielding Middleware" before being handed to the Executor.
* **Observability:** Integrated with the "Agent Chain Tracer" to visualize advisor interdictions.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
