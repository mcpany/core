# Design Doc: Environment Sovereignty Enforcer (ESE)
**Status:** Draft
**Created:** 2026-06-10

## 1. Context and Scope
The "Identity Leakage via Process Environment" (ILPE) vulnerability proves that hardware-attested mission-root identity tokens can be leaked through process environment blocks and temporary file metadata during tool execution. This allows low-trust tools to hijack the lineage of a high-trust agent.

The **Environment Sovereignty Enforcer (ESE)** is a core security service for the Ephemeral Privilege Manager (EPM) and LOWA providers. It mandates hardware-attested "Environment Scrubbing" for all tool-executing specialists, ensuring that mission-root identity tokens are isolated and never leaked to the host environment.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Mandate hardware-attested "Environment Scrubbing" for all tool calls.
    *   Isolate mission-root identity tokens from the process environment.
    *   Neutralize ILPE exfiltration via temporary metadata.
    *   Support both local and containerized execution environments.
*   **Non-Goals:**
    *   Replacing the entire OS process isolation.
    *   Managing non-security-critical environment variables (e.g., `PATH`).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Execute a high-privilege Python script without leaking the hardware-attested session token to the host's `/proc` or temporary file system.
*   **The Happy Path (Tasks):**
    1.  The agent requests execution of a privileged script via a Command Adapter.
    2.  The EPM requests an environment block from the ESE.
    3.  ESE generates a scrubbed, session-bound environment that excludes the mission-root identity token.
    4.  The script is executed in a hardware-attested, isolated environment.
    5.  Upon completion, ESE performs a secure "Post-Flight Wipe" of temporary file metadata.
    6.  Mission-root identity remains sovereign and un-leaked.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        EPM[Ephemeral Privilege Manager] -->|Request Env| ESE[Environment Sovereignty Enforcer]
        ESE -->|Scrub & Isolate| EnvBlock[Isolated Env Block]
        EnvBlock -->|Execute| CommandAdapter[Command Adapter]
        CommandAdapter -->|Post-Flight| ESE
        ESE -->|Secure Wipe| TempFS[Temporary Metadata]
    ```
*   **APIs / Interfaces:**
    *   `ScrubEnvironment(baseEnv []string) ([]string, error)`: Removes sensitive identity tokens from the environment.
    *   `WipeMetadata(taskID string) error`: Performs secure deletion of temporary task metadata.
*   **Data Storage/State:**
    *   Maintains a temporary, hardware-locked manifest of all task-bound environment blocks to ensure secure wiping.

## 5. Alternatives Considered
*   **Container Isolation Only (Docker):** Rejected because many local agents (Claude Code, Gemini CLI) often execute commands on the host directly, and even containers can leak metadata through mounted volumes.
*   **Environment Var Encryption:** Rejected because the final tool execution often requires raw strings; ESE focuses on *removing* the tokens from the environment altogether when they are not explicitly required by the tool.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** ESE requires hardware attestation (TPM) to perform its final wipe and attestation signals.
*   **Observability:** Integrated with the `Audit Log`, showing "Environment Sanitization" events.

## 7. Evolutionary Changelog
*   **2026-06-10:** Initial Document Creation.
*   **2026-06-17:** **Addressing ILPE (Identity Leakage via Process Environment)**
    *   **Context:** Market sync revealed a new exploit pattern where mission-root identity tokens are exfiltrated via standard process environment blocks.
    *   **Architecture Adjustment:** Mandating **Hardware-Attested Environment Scrubbing** in Section 4. All tool-executing specialists are now required to provide a TPM-signed proof of environment purity before receiving any session-bound capability.
    *   **Security Impact:** Mitigates the risk of mission-root token exfiltration by malicious subagents or un-attested tools.
