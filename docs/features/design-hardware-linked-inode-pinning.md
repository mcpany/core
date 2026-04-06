# Design Doc: Hardware-Linked Inode Pinning (HLIP)
**Status:** Draft
**Created:** 2026-04-06

## 1. Context and Scope
As AI agents increasingly rely on project-local configuration files (e.g., `.mcpany/settings.json`, `.claude/settings.json`), a new class of TOCTOU (Time-of-Check to Time-of-Use) attacks has emerged. In these attacks, a malicious process or subagent swaps a validated configuration file with a compromised one after the initial security check but before the file is read by the agent.

Hardware-Linked Inode Pinning (HLIP) addresses this by cryptographically binding a file's hardware Inode to the active session. This ensures that the file being read is exactly the same physical resource that was verified, neutralizing any filesystem-level racing attempts.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound file handle persistence for project-local configurations.
    * Detect and block unauthorized file swaps (SIR exploits) during an active reasoning session.
    * Provide TPM-backed attestation for pinned file handles.
* **Non-Goals:**
    * Securing remote filesystem mounts (restricted to local hardware-bound volumes).
    * Managing file permissions (handled by the OS and DPG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect for AI Swarms
* **Primary Goal:** Ensure that a validated project configuration cannot be swapped during the agent's execution.
* **The Happy Path (Tasks):**
    1. MCP Any validates a project-local configuration file.
    2. The HLIP middleware retrieves the file's hardware Inode and binds it to a session-specific TPM key.
    3. A malicious process attempts to swap the file using a symlink race.
    4. HLIP detects the Inode mismatch during the subsequent read operation.
    5. The tool call is interdicted, and a security alert is issued.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Config Load Request] --> B[Path Normalization]
        B --> C[Initial Validation]
        C -- Success --> D[HLIP Inode Binding]
        D --> E[TPM Signature Generation]
        E --> F[Pinned File Handle Storage]
        G[Subsequent Read] --> H[Inode Match Check]
        H -- Mismatch --> I[Interdict & Alert]
        H -- Match --> J[Allow Execution]
    ```
* **APIs / Interfaces:**
    * `hlip.PinFile(path) -> HandleID`: Validates and locks an Inode.
    * `hlip.VerifyHandle(handleID) -> bool`: Checks current Inode against the pinned signature.
* **Data Storage/State:**
    * **TPM Storage:** Secure storage for session-bound Inode signatures.

## 5. Alternatives Considered
* **Continuous Path Re-validation:** Rejected due to the inherent race condition between path resolution and file open.
* **Kernel-level File Locking (flock):** Rejected as it can be bypassed by unprivileged processes or symlink swaps at the directory level.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HLIP is a mandatory component of the Pre-Flight Sandbox Validator.
* **Observability:** Integrated with the "Inode Security Monitor" for real-time visualization of blocked swaps.

## 7. Evolutionary Changelog
* **2026-04-06:** Initial Document Creation.
