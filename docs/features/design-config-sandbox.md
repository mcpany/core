# Design Doc: Zero-Trust Config Sandbox

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
The "Claude Code RCE" exploit of early 2026 demonstrated that AI agents are vulnerable to malicious repository-level configurations. If an agent (like Claude Code or OpenClaw) automatically loads configuration from a `.claudecode` or `.mcpany` file in a cloned repository, an attacker can inject malicious tool definitions or environment variables. MCP Any must protect itself and the agents it serves by sandboxing or verifying all non-system configurations.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent automatic execution of tools defined in repository-local configuration files without human approval or cryptographic verification.
    * Implement a "Config Signature" verification loop using developer public keys.
    * Sandbox environment variables loaded from repo-local files to prevent leakage of system-level secrets.
* **Non-Goals:**
    * Providing a full VM sandbox for tool execution (this is handled at the OS/Container level).
    * Validating the *logic* of the code within tools (this is handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer
* **Primary Goal:** Clone a public repository and use MCP Any to explore its tools without risking RCE.
* **The Happy Path (Tasks):**
    1. User `cd`s into a newly cloned repository containing a `.mcpany/config.yaml`.
    2. User runs `mcpany start`.
    3. MCP Any detects the local config and checks for a `signature.asc`.
    4. If unsigned, MCP Any prompts the user: "Untrusted configuration detected in [path]. Review and approve? [Y/n]".
    5. Once approved (or if signed by a trusted key), the tools are loaded into a restricted "Repo-Scope" context.

## 4. Design & Architecture
* **System Flow:**
    - **Config Watcher**: Monitors the current working directory for `.mcpany` files.
    - **Trust Manager**: Maintains a keyring of trusted developer identities (GPG/SSH keys).
    - **Approval Store**: Records hashes of manually approved configurations to prevent repeated prompts.
* **APIs / Interfaces:**
    - `mcpany config trust [key-id]`: Adds a developer key to the trusted list.
    - `mcpany config verify`: Manually triggers a verification of the local config.
* **Data Storage/State:**
    - `~/.mcpany/trust_store.json`: Stores trusted keys and approved config hashes.

## 5. Alternatives Considered
* **Global Config Only**: Only allow tools to be defined in the user's home directory. *Rejected* because it breaks the portability of agent-ready repositories.
* **Automatic Dry-Run**: Run all new tools in a dry-run mode first. *Rejected* as it doesn't prevent environment variable theft.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This feature specifically addresses "Trusting the Environment" risks.
* **Observability:** Logs must clearly distinguish between "System Tools" and "Untrusted Repo Tools."

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
