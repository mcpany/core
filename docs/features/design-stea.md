# Design Doc: Startup-Time Environment Attestor (STEA)
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
The disclosure of CVE-2026-22177 (OpenClaw Startup-Time Env Var Injection) has highlighted a critical vulnerability in autonomous agent frameworks: the ability for an attacker to compromise the agent *before* it even begins processing its first task. By injecting malicious environment variables (e.g., `LD_PRELOAD`, `PYTHONPATH`) at server startup, attackers can gain early-stage RCE that bypasses post-boot security filters.

The **Startup-Time Environment Attestor (STEA)** is the authoritative "Boot Guardian" for MCP Any. It ensures that the execution environment is verified and hardware-attested before the gateway enters an operational state.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform hardware-attested validation of all environment variables and startup flags.
    * Mandate a "Verified Baseline" for sensitive variables (e.g., `PATH`, `HOME`).
    * Block startup if any unauthorized or high-risk injection pattern is detected.
    * Provide a signed "Boot Integrity Manifest" for downstream agents.
* **Non-Goals:**
    * Managing environment variables changed *after* startup (handled by Continuous CPCP Enforcer).
    * Replacing OS-level container isolation.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Admin
* **Primary Goal:** Prevent a CI/CD-deployed MCP Any instance from being hijacked via a poisoned `.env` file or orchestrator-level injection.
* **The Happy Path (Tasks):**
    1. The MCP Any process is initiated by the system orchestrator.
    2. STEA intercepts the startup sequence before any service initialization.
    3. STEA retrieves the hardware-attested "Baseline Policy" from the TPM.
    4. STEA compares the current `os.Environ()` against the policy.
    5. STEA detects an unauthorized `LD_PRELOAD` injection.
    6. STEA terminates the process immediately and logs a "Boot Tamper Alert."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Init[Process Init] -->|Intercept| STEA[STEA]
        STEA -->|Fetch| TPM[Hardware Baseline]
        STEA -->|Scan| Env[os.Environ]
        Env -->|Validation| Filter{Baseline Check}
        Filter -->|Fail| Terminate[Exit 1 & Alert]
        Filter -->|Pass| Sign[Generate Boot Manifest]
        Sign -->|Operation| Gateway[Operational Gateway]
    ```
* **APIs / Interfaces:**
    * `internal/stea.VerifyBootEnv()`: Core validation function called in `main.go`.
    * `Policy: boot_allow_list`: Regex patterns for authorized environment keys.
* **Data Storage/State:**
    * Hardware-attested baseline stored in the host TPM/Secure Enclave.

## 5. Alternatives Considered
* **Static Binary Only:** Rejected because flexible configuration (environment variables) is a core requirement for cross-platform deployment.
* **Post-Boot Scanning:** Rejected because by the time scanning occurs, a malicious `LD_PRELOAD` could have already hijacked the process memory.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** STEA is the "Root of Trust" for the entire operational session.
* **Observability:** Logs "Boot Integrity" status to the system secure log.

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
