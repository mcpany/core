# Design Doc: Obfuscated Execution Guard

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw, there is an increasing risk of "Obfuscated Command Injection." Malicious actors or misaligned agents may attempt to execute dangerous shell commands by encoding them in Base64, Hex, or using complex shell expansions (e.g., `$(echo 'Y3VybCAuLi4=' | base64 -d)`). MCP Any needs a native middleware to detect, intercept, and require human approval for such patterns.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Detect common obfuscation patterns (Base64, Hex, URL encoding) in tool arguments.
    *   Intercept calls containing suspected obfuscation and trigger a "Human-in-the-Loop" (HITL) pause.
    *   Provide a "De-obfuscated Preview" to the user for informed approval.
    *   Support configurable sensitivity levels for different environments.
*   **Non-Goals:**
    *   Perfectly decyphering every possible shell trick (impossible).
    *   Replacing the Policy Firewall (this is a specialized safety layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer using a local agent swarm.
*   **Primary Goal:** Prevent an agent from accidentally or maliciously executing a hidden script that exfiltrates local data.
*   **The Happy Path (Tasks):**
    1.  User enables `ObfuscatedExecutionGuard` in `mcpany.yaml`.
    2.  An agent tries to call a shell tool with an argument like `sh -c "$(echo ZWNobyAiSGFja2VkISIK | base64 -d)"`.
    3.  MCP Any middleware detects the Base64 pattern and the `base64 -d` pipe.
    4.  The tool execution is paused.
    5.  The UI displays a warning: "Suspicious Obfuscated Command Detected."
    6.  The UI shows the de-obfuscated intent: `echo "Hacked!"`.
    7.  The user rejects the execution.

## 4. Design & Architecture
*   **System Flow:**
    - **Heuristic Engine**: Scans all incoming `tools/call` arguments for regex patterns matching Base64, Hex strings, and shell redirection.
    - **Entropy Analysis**: Measures string entropy to detect suspicious non-natural language blobs.
    - **HITL Integration**: If a threshold is met, the call is pushed to the HITL Pending Queue.
*   **APIs / Interfaces:**
    - Middleware hook in the execution pipeline.
    - Notification event sent to the UI via WebSocket.
*   **Data Storage/State:** Transient state in the HITL manager; audited logs in the persistent database.

## 5. Alternatives Considered
*   **Static Blocklist**: Blocking all shell-related tools. *Rejected* as it breaks legitimate agent workflows.
*   **LLM-based Detection**: Using another LLM to scan for maliciousness. *Rejected* due to latency and the risk of the "evaluator" being bypassed by the same tricks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The guard runs *before* the tool is ever touched by the OS.
*   **Observability:** All blocked/paused attempts are logged with high-fidelity traces for forensic analysis.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
