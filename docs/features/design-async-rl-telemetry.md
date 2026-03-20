# Design Doc: Async RL Telemetry Orchestrator

**Status:** Draft
**Created:** 2026-05-11

## 1. Context and Scope
OpenClaw-RL v1.0 and recent research into "Process-Reward Models" (PRM) demand that agent infrastructure supports high-frequency telemetry export for real-time training. Traditional blocking telemetry adds unacceptable latency to agent reasoning. The Async RL Telemetry Orchestrator provides a non-blocking, high-speed bridge for exporting reasoning traces, tool performance, and feedback tokens to RL training pipelines.

## 2. Goals & Non-Goals
* **Goals:**
    * Collect reasoning traces and multi-modal feedback asynchronously from all connected agents.
    * Provide a low-overhead interface for agents to signal "Reward Events" (e.g., successful task completion).
    * Export normalized telemetry to external RL training backends (e.g., OpenClaw-RL).
    * Ensure privacy by semantically sanitizing traces before export.
* **Non-Goals:**
    * Managing the RL training process itself.
    * Storing high-volume traces locally (traces should be streamed to the collector).

## 3. Critical User Journey (CUJ)
* **User Persona:** AI Researcher / RL Engineer
* **Primary Goal:** Collect real-time performance data from a fleet of agents to optimize a task-specific policy.
* **The Happy Path (Tasks):**
    1. The researcher configures an "RL Pipeline" endpoint in MCP Any.
    2. Agents performing tasks send reasoning deltas and tool results to the Orchestrator.
    3. The Orchestrator buffers and batches the data to minimize network overhead.
    4. Data is sanitized for PII and exported to the RL endpoint.
    5. The policy is optimized in the background based on the aggregated "Rollouts."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|Trace Delta| Buffer[Ring Buffer]
        Buffer -->|Batch| Sanitizer[Semantic Sanitizer]
        Sanitizer -->|Stream| RLEndpoint[RL Training Endpoint]
    ```
* **APIs / Interfaces:**
    * `SubmitTrace(agent_id, trace_fragment)`: Non-blocking trace submission.
    * `ReportReward(mission_id, score, context)`: Signals a policy-improvement event.
* **Data Storage/State:** High-speed in-memory ring buffers. No disk persistence for traces.

## 5. Alternatives Considered
* **Direct Agent-to-RL Export**: Rejected because it increases agent complexity and reasoning latency.
* **Log-Based Post-Processing**: Rejected because RL requires "In-Session" feedback loops for high-frequency optimization.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Traces must be sanitized of secrets and PII before export. Only authorized RL pipelines can subscribe to the telemetry stream.
* **Observability**: A "Policy Drift" dashboard in the UI visualizes the effectiveness of the feedback loop.

## 7. Evolutionary Changelog
* **2026-05-11:** Initial Document Creation.
