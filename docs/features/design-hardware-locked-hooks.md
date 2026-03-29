# Design Doc: Hardware-Locked Configuration & Hook Anchoring (HLCA)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
Recent critical vulnerabilities in Claude Code (CVE-2025-59536) and Gemini CLI have exposed a massive security gap: configuration files are now active execution layers. Attackers can embed malicious lifecycle hooks or override base URLs in repository-level settings, leading to Remote Code Execution (RCE) and credential theft upon simply opening a project.

MCP Any must solve this by decoupling project configuration from execution authority. The "Universal Agent Bus" must ensure that no hook or security-sensitive setting is active unless it has been explicitly anchored to a hardware-attested user session.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind `.claude/settings.json`, `GEMINI.md`, and other config files to a TPM-signed user session.
    * Mandate sandboxed behavioral profiling for all automated hooks (HSE).
    * Provide a "Hardware Attestation Receipt" for every authorized configuration state.
    * Neutralize TOCTOU (Time-of-Check to Time-of-Use) attacks on configuration files.
* **Non-Goals:**
    * Replacing existing config formats (JSON, YAML, MD).
    * Managing the LLM's internal reasoning logic (this is handled by other layers like ARI/SRM).
    * Providing a general-purpose secret manager (though it integrates with NIM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / Enterprise AI Architect
* **Primary Goal:** Securely open an untrusted open-source repository without risking RCE via malicious hooks.
* **The Happy Path (Tasks):**
    1. User clones an untrusted repository.
    2. MCP Any detects a project-level configuration file containing executable hooks.
    3. MCP Any moves the configuration into a "Quarantine State" and notifies the user.
    4. The Hook Sovereignty Enforcer (HSE) performs a sandboxed dry-run and static analysis.
    5. User reviews the profiling report (e.g., "This hook attempts to access `~/.ssh/`").
    6. User approves the specific hooks via a hardware-bound MFA (TouchID/TPM).
    7. MCP Any generates an HLCA "Anchor" (a signed hash of the config bound to the session).
    8. The agent is allowed to load the configuration, with MCP Any enforcing that the file's Inode matches the Anchor.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant FS as Filesystem
        participant HLCA as HLCA Middleware
        participant HSE as Hook Sovereignty Enforcer
        participant TPM as Hardware TPM/SEP
        participant Agent as AI Agent Runtime

        FS->>HLCA: Config File Detected
        HLCA->>HSE: Profile Hooks
        HSE-->>HLCA: Safety Report
        HLCA->>User: Request MFA Approval
        User->>TPM: Sign Approval
        TPM-->>HLCA: Attestation Token
        HLCA->>FS: Pin Inode + Hash
        HLCA->>Agent: Expose "Attested View"
    ```
* **APIs / Interfaces:**
    * `POST /v1/attestation/anchor`: Anchors a specific configuration file.
    * `GET /v1/attestation/status`: Checks if a project's state is anchored.
    * `hook_profile`: Internal interface for HSE static analysis results.
* **Data Storage/State:**
    * HLCA maintains a local SQLite state db (Blackboard) containing `{path, inode, sha256, tpm_signature, expiry}`.

## 5. Alternatives Considered
* **Binary Allow-listing:** Rejected because it doesn't handle dynamic configuration changes or repo-specific hooks well.
* **Static Sandboxing (Docker):** Necessary but insufficient. Docker protects the host but doesn't prevent "Context-Splicing" or "Identity Theft" within the agent's virtual environment. HLCA provides the "Identity-to-Config" binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All configuration is "Untrusted" until hardware-anchored. Even a 1-byte change in the file breaks the anchor.
* **Observability:** HLCA logs all attestation events to the Local Security Audit Log, including the full diff of what was approved.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
