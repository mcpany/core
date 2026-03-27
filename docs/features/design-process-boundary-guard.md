# Design Doc: Process Boundary Guard
**Status:** Draft
**Created:** 2026-03-27

## 1. Context and Scope
The disclosure of CVE-2026-22177 (OpenClaw RCE) has highlighted a critical vulnerability where malicious project-local configurations can inject dangerous process-control environment variables (e.g., `NODE_OPTIONS`, `LD_PRELOAD`, `PYTHONPATH`) into the agent runtime. This allows attackers to achieve arbitrary code execution even before any tools are invoked. The Process Boundary Guard acts as a "Sanitizing Proxy" for the execution environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically scrub a curated list of dangerous process-control variables from all project-local and tool-specific configurations.
    * Enforce an environment variable allowlist for high-trust tool execution.
    * Log and alert on attempts to inject restricted variables.
    * Provide a mechanism for explicit user attestation of necessary but dangerous variables.
* **Non-Goals:**
    * Managing standard non-privileged environment variables (e.g., `LOG_LEVEL`).
    * Replacing kernel-level isolation (e.g., namespaces/cgroups).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious CI/CD Engineer
* **Primary Goal:** Prevent a malicious repository from hijacking the agent runtime via `.claude/settings.json`.
* **The Happy Path (Tasks):**
    1. A malicious repository is cloned containing a configuration that sets `NODE_OPTIONS='--require /tmp/malware.js'`.
    2. The agent attempts to load this configuration via MCP Any.
    3. The Process Boundary Guard intercepts the configuration load.
    4. The Guard identifies `NODE_OPTIONS` as a restricted process-control variable.
    5. The variable is scrubbed from the environment map before being passed to the agent runtime.
    6. A security warning is surfaced to the user in the MCP Any dashboard.

## 4. Design & Architecture
* **System Flow:**
    * **Ingestion Layer**: Intercepts configuration reads from filesystem or API.
    * **Scrubbing Engine**: Compares the environment map against a `RestrictedVariableRegistry`.
    * **Attestation Store**: Stores user-signed exceptions for specific variables/tools.
    * **Enforcement Point**: Injects the sanitized environment into the tool execution or subagent spawn process.
* **APIs / Interfaces:**
    * `SanitizeEnvironment(envMap) -> sanitizedMap`
    * `RegisterRestrictedVariable(varName, riskLevel)`
    * `ApproveVariableException(missionId, varName, toolId)`
* **Data Storage/State:** Persistent storage for user-attested exceptions and a built-in registry of known dangerous variables.

## 5. Alternatives Considered
* **Runtime Scanning (e.g., eBPF)**: Rejected as the primary mechanism due to complexity and the need for early-stage (pre-spawn) mitigation, though it is a valuable complementary layer.
* **Manual Review Only**: Rejected as it fails against automated swarm executions and high-frequency "clone-and-run" workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements the principle of least privilege for the execution environment.
* **Observability:** Provides detailed logs of scrubbed variables and injection attempts for forensic analysis.

## 7. Evolutionary Changelog
* **2026-03-27:** Initial Document Creation.
