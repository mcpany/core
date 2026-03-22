# Design Doc: Pre-Loading Configuration Attestation (PLCA)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
Recent vulnerabilities in major agent frameworks (e.g., CVE-2026-33068 in Claude Code) have demonstrated that the configuration loading process itself is a critical attack vector. If an agent loads untrusted project-local settings before the security gatekeeper is initialized, the entire sandbox can be bypassed. PLCA aims to solve this by mandating a hardware-attested validation of the environment *before* any configuration parsing occurs.

## 2. Goals & Non-Goals
* **Goals:**
    * Ensure the environment is in a verified state before reading any `.mcpany`, `.claude`, or `GEMINI.md` files.
    * Use TPM-bound signatures to verify that project-local configurations haven't been tampered with since the last user attestation.
    * Provide an "External Gatekeeper" that operates independently of the agent's internal state.
* **Non-Goals:**
    * It will not sanitize the *content* of the configuration (this is handled by the Policy Firewall).
    * It will not manage agent execution itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Administrator
* **Primary Goal:** Prevent a malicious repository from hijacking the local AI agent via a poisoned configuration file.
* **The Happy Path (Tasks):**
    1. User clones a repository.
    2. User runs `mcpany start` or an agent-integrated command.
    3. MCP Any intercepts the boot sequence.
    4. PLCA generates a hardware-attested snapshot of all project-local config files.
    5. PLCA compares hashes against the local "Trusted Manifest."
    6. If a mismatch or new file is found, the boot is halted and the user is prompted for attestation.
    7. Once verified, PLCA releases the "Loading Lock," allowing the agent to parse the configs.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Boot Trigger] --> B[PLCA Interceptor]
        B --> C{Verify Hardware Root}
        C -- Success --> D[Generate Config Hashes]
        D --> E{Match Trusted Manifest?}
        E -- Yes --> F[Release Loading Lock]
        E -- No --> G[Trigger HITL Re-Attestation]
        G -- Approved --> H[Update Manifest & Release]
        G -- Denied --> I[Halt Execution]
    ```
* **APIs / Interfaces:**
    * `Internal-Gatekeeper-Protocol`: A low-level binary interface used to pause/resume the agent's file-system access during boot.
* **Data Storage/State:**
    * `trusted_manifest.db`: A TPM-encrypted local database storing hashes of verified configuration blocks.

## 5. Alternatives Considered
* **Content Sandboxing:** Relying on the agent to ignore dangerous fields. Rejected because it's vulnerable to logic flaws in the agent's parser (as seen in CVE-2026-33068).
* **Path-Based Deny-Lists:** Blocking specific filenames. Rejected because attackers can use new, unlisted filenames that are then "discovered" by the agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PLCA must be the first process to touch the project directory. Any race condition where the agent reads before PLCA locks is a failure.
* **Observability:** Log every config file hash and its attestation status to the `audit_log.db`.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
