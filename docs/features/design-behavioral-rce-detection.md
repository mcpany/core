# Design Doc: Behavioral RCE Detection Middleware
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
With the rise of autonomous agents, "One-Click RCE" has become the primary threat vector. Attackers exploit the agent's ability to execute tools by triggering a sequence of actions that, individually, might seem benign or common (e.g., `git clone`, `ls`, `chmod`, `exec`). Static analysis of single tool calls or configuration files cannot catch these distributed attack chains. MCP Any needs a stateful middleware that monitors the behavioral patterns of tool call sequences in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a stateful "Behavioral Tracer" that tracks sequences of tool calls per-agent session.
    * Define and detect "Suspicious Attack Patterns" (e.g., Download -> Permit -> Execute).
    * Provide a "Quarantine" mechanism that suspends a session and requires manual human intervention when a pattern is matched.
    * Minimal latency impact on tool execution.
* **Non-Goals:**
    * Predicting all possible malicious intents (focus is on known RCE patterns).
    * Replacing existing static policy engines (this is an additional layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an agent to explore a new open-source library.
* **Primary Goal:** Prevent the agent from being tricked into running a malicious script hidden in the library's setup process.
* **The Happy Path (Tasks):**
    1. Agent calls `http_get` to download a "setup.sh" script.
    2. Agent calls `chmod` to make it executable.
    3. Behavioral Engine flags the `chmod` on a freshly downloaded file as "Suspicious Step 2".
    4. Agent calls `shell_exec` on the script.
    5. Behavioral Engine matches the "Download -> Permit -> Execute" pattern and immediately suspends the tool call.
    6. User receives an urgent notification: "Suspicious RCE Pattern Detected: Automated execution of a newly downloaded file."

## 4. Design & Architecture
* **System Flow:**
    `Tool Request` -> `Policy Firewall` -> `Behavioral Middleware` -> `Upstream MCP Server`
    1. **Observation**: Every tool call is logged to a short-term, in-memory sliding window buffer (`Sequence Buffer`).
    2. **Pattern Matching**: The buffer is matched against a set of regex-like "Behavioral Signatures."
    3. **Scoring**: Each step in a signature increases the "Risk Score" of the session.
    4. **Intervention**: If the score exceeds the threshold, the `HITL Middleware` is triggered to pause execution.
* **APIs / Interfaces:**
    * `Internal Interface`: `BehavioralTracer.Observe(ToolCall)`
    * `Signature Schema`: A JSON/YAML format for defining multi-step tool patterns.
* **Data Storage/State:**
    * `Redis/In-Memory Map`: Stores the last `N` tool calls for each active `session_id`.

## 5. Alternatives Considered
* **Static Sandboxing**: While effective, it often breaks legitimate workflows (e.g., a dev actually wanting to run a script). Behavioral analysis allows for more nuance.
* **Model-based Analysis**: Using an LLM to "reason" if a sequence is safe. Rejected as a primary mechanism due to latency and the risk of the "Analyzer LLM" being bypassed by the same prompt injection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The engine itself must be immutable and run in the core gateway process.
* **Observability**: Patterns matched are logged with high-priority "Security Alert" status.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
