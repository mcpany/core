# Design Doc: Speculative Execution Guard
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
As agent swarms become more complex and multi-layered (e.g., Gemini's CDQ model), the latency introduced by security quorums and background attestation has become a primary bottleneck for user experience. To remain the high-performance bus for agentic coordination, MCP Any must support "Speculative Execution."

However, speculative execution introduces the risk of "Branch Contamination" and "Hallucinatory Context," where results from an unverified (and potentially malicious or incorrect) tool call are ingested into the agent's reasoning loop. The Speculative Execution Guard provides a transactional middleware layer to manage this risk.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable agents to speculatively execute low-risk (Read-Only) tools while attestation quorums run in the background.
    * Provide a "Shadow State" buffer to hold speculative results in isolation.
    * Ensure atomicity of state commits: results are only merged into the global Blackboard/Context upon successful attestation.
    * Facilitate automatic rollback of agent state and purging of buffers upon attestation failure.
* **Non-Goals:**
    * Will not support speculative execution for high-risk (Write/Execute) tools (e.g., `run_shell_command`).
    * Will not perform the actual attestation; relies on external quorums (CDQ/MAQ).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Agent Swarm Orchestrator
* **Primary Goal:** Execute a sequence of discovery and tool-call steps with sub-millisecond latency despite multi-agent security quorums.
* **The Happy Path (Tasks):**
    1. Agent initiates a tool call marked for speculative execution.
    2. Speculative Execution Guard intercepts the call and creates a "Shadow Branch" in the session state.
    3. Tool is executed in the background; results are returned to the agent and stored in the "Shadow Buffer."
    4. Agent continues reasoning based on the speculative results.
    5. Background Discovery Quorum (CDQ) successfully attests to the tool's safety.
    6. Guard commits the "Shadow Buffer" to the global mission root state and marks the branch as "Verified."

## 4. Design & Architecture
* **System Flow:**
    [Agent] -> [Speculative Execution Guard] -> (Shadow Branch Created) -> [Tool]
    (Attestation running in parallel)
    [Quorum] -> (Success) -> [Speculative Execution Guard] -> (Commit to Blackboard)
    [Quorum] -> (Failure) -> [Speculative Execution Guard] -> (Purge Buffer & Rollback Agent)
* **APIs / Interfaces:**
    * `speculative: true` flag in tool-call headers.
    * `onAttestationSuccess(branchId)` and `onAttestationFailure(branchId)` callbacks for quorum integration.
* **Data Storage/State:**
    * Uses a copy-on-write (COW) model for the mission-root state fragments.
    * Speculative results are stored in an ephemeral, memory-mapped buffer linked to the specific `branchId`.

## 5. Alternatives Considered
* **Blocking Execution:** Rejected due to prohibitive UX latency (2s+ per call).
* **Direct-to-Blackboard Writing:** Rejected due to "Branch Contamination" risks identified in OpenClaw post-mortems.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative branches are isolated; high-risk tools are excluded from speculative paths.
* **Observability:** Visualized via the "Speculative State Inspector" in the UI, showing buffer status and commit/rollback events.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
