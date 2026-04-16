# Design Doc: Risk-Adaptive Auto-Execution Gate
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents evolve into autonomous teammates, "Approval Fatigue" has become a primary bottleneck for user productivity. Requiring manual user attestation for every read-only or low-risk tool call (e.g., `ls`, `cat`, `grep`) disrupts the reasoning loop. Conversely, the "dangerously-skip-permissions" pattern exposes systems to catastrophic RCE via structural metadata poisoning or prompt injection.

The Risk-Adaptive Auto-Execution Gate is a high-speed security middleware that sits between the agent reasoning loop and the MCP tool execution layer. It leverages transcript classification and real-time injection probes to provide a "middle path": automatically authorizing safe actions while escalating high-privilege or high-risk tool calls for hardware-attested user approval.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce manual approval frequency by at least 80% for common development workflows.
    *   Automatically interdict tool calls that exhibit prompt or command injection patterns.
    *   Integrate with the Injection-Shielding Middleware for argument-level semantic validation.
    *   Provide a fallback for "Suspicious" actions to hardware-attested user reviews.
*   **Non-Goals:**
    *   Replacing the Policy Firewall (Rego/CEL) which handles static access control.
    *   Managing local process isolation (delegated to gVisor/Docker).
    *   Automating high-risk actions (e.g., `rm -rf /`, `ssh-add`) without explicit user opt-in.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Senior Software Engineer using an autonomous coding agent.
*   **Primary Goal:** Allow the agent to perform deep repository analysis (read-only) without constant interruptions while ensuring any code execution is verified.
*   **The Happy Path (Tasks):**
    1.  The User enables "Auto-Mode" with a mission-root constraint (e.g., "Safe for Read-Only").
    2.  The Agent requests `ls src/components` to understand the project structure.
    3.  The Gate classifies the request as "Low Risk" (Read-Only + Standard Path).
    4.  The Gate performs an injection probe on the arguments; no shell metacharacters found.
    5.  The action is executed automatically; the User sees a "Passive Audit" notification.
    6.  The Agent later requests `npm install --save malicious-package` (via a poisoned `README.md` instruction).
    7.  The Gate classifies the request as "High Risk" (Filesystem Write + External Network).
    8.  The Gate pauses execution and surfaces an MFA Attestation Dialog to the User.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Agent Reasoning Engine] -->|Tool Call Request| B(Risk-Adaptive Gate)
        B --> C{Risk Classifier}
        C -->|Low Risk / Safe| D[Injection Shielding Probe]
        C -->|High Risk / Suspicious| E[MFA Attestation UI]
        D -->|Pass| F[MCP Tool Execution]
        D -->|Fail| G[Interdiction Alert]
        E -->|User Approved| F
        E -->|User Denied| G
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/security/evaluate-risk`: Evaluates a proposed tool call and returns a `Decision` (Auto-Execute, Approve, Block).
    *   `Decision.Reasoning`: A semantic explanation of why the risk score was assigned (e.g., "Read-only access to allowlisted directory").
*   **Data Storage/State:**
    *   **Transcript Buffer**: Short-term memory of recent tool calls to detect "low-and-slow" exfiltration patterns.
    *   **Risk Profile Registry**: A persistent store of tool-specific risk weights and historically approved allowlists.

## 5. Alternatives Considered
*   **Static Allowlisting**: Rejected because path-based allowlists are brittle and cannot detect prompt-injection within valid paths (e.g., `cat "file; rm -rf /"`).
*   **Always-HITL**: Current state. Rejected due to the "Approval Fatigue" bottleneck identified in Claude Code v4.7 market shift.
*   **Pure LLM-based Safety Checks**: Rejected due to latency and the risk of the "Safety Agent" itself being subverted by the same prompt injection.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The gate operates in a "Deny-by-Default" mode. If the risk classifier returns a confidence score below the threshold, it MUST escalate to HITL.
*   **Observability:** Every "Auto-Execute" decision is logged to the Local Security Audit Log with the associated transcript fragment for post-hoc review.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
