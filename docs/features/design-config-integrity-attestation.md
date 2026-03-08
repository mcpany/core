# Design Doc: Config Integrity Attestation
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As AI agents (e.g., Claude Code, OpenClaw) increasingly rely on project-local configuration files (`.mcp.json`, `.claude/settings.json`, `.env`), these files have become a primary attack vector for Remote Code Execution (RCE) and credential exfiltration. Malicious actors can commit poisoned configurations that execute unauthorized "hooks" or redirect API traffic to malicious endpoints. MCP Any needs a robust mechanism to ensure that any configuration it ingests from a local project directory is authorized and untampered.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically verify the integrity of local configuration files before loading.
    * Provide a "Manual Approval" flow for untrusted or modified configurations.
    * Prevent "Hook Injection" attacks by strictly validating shell command hooks against a known-good signature.
    * Maintain a local "Trust Database" of approved project configurations.
* **Non-Goals:**
    * Encrypting local configuration files (focus is on integrity and authorization, not secrecy at rest).
    * Managing remote/cloud-based configurations (scope is limited to local project-bound configs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely open a community-contributed repository and use MCP Any tools without risking RCE from malicious project configs.
* **The Happy Path (Tasks):**
    1. User clones a repository containing a `.mcpany.yaml` config.
    2. User runs `mcpany serve` in the repository root.
    3. MCP Any detects the new configuration and checks its local Trust Database.
    4. Finding no record, MCP Any pauses and prompts the user (via CLI or UI) to review the configuration and its signatures.
    5. User reviews the changes (specifically any `hooks` or `overrides`) and selects "Approve and Sign."
    6. MCP Any generates a local HMAC signature of the file and stores it in the Trust Database.
    7. Future starts of the server in this directory proceed automatically as long as the file hash matches the signature.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Start MCP Any] --> B{Config in Project?}
        B -- No --> C[Load Global Config]
        B -- Yes --> D{In Trust DB?}
        D -- Yes --> E{Hash Match?}
        E -- Yes --> F[Load Config]
        E -- No --> G[Quarantine & Prompt User]
        D -- No --> G
        G -- User Approves --> H[Sign & Add to Trust DB]
        H --> F
        G -- User Rejects --> I[Abort/Load Global Only]
    ```
* **APIs / Interfaces:**
    * `mcpany trust sign <file>`: CLI command to manually sign a config file.
    * `GET /api/v1/trust/status`: Check the trust status of the current project.
    * `POST /api/v1/trust/approve`: Approve a quarantined configuration.
* **Data Storage/State:**
    * `~/.mcpany/trust.db`: A local SQLite database storing file paths, hashes, and user approval timestamps.

## 5. Alternatives Considered
* **Purely Declarative Allowlist**: Rejected because it's too rigid for developers who frequently tweak configurations.
* **OS-Level File Locking**: Rejected as it doesn't prevent malicious commits from being pulled into a collaborator's environment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Human-in-the-Loop" approval is the primary defense. The signature ensures that even if a file is modified via `git pull`, it must be re-approved.
* **Observability:** All trust decisions (approvals, rejections, hash mismatches) are logged to the Audit Log with high severity.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
