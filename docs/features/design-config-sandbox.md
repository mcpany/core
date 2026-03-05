# Design Doc: Config Sandbox Validator

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The recent CVE-2025-59536 (Claude Code RCE) demonstrated that project-level configuration files (like `.claudecode.json` or equivalent) are a high-value target for attackers. By placing a malicious config in a repo, an attacker can execute arbitrary commands when a user opens that repo with an AI agent.

MCP Any must treat all "Project-Level" configurations as untrusted and validate them in a secure sandbox before merging them into the active runtime.

## 2. Goals & Non-Goals
* **Goals:**
    * Parse `.mcpany/config.yaml` from project roots in a restricted parser.
    * Validate all "Hooks" and "Commands" against a strict whitelist.
    * Require explicit User Approval (HITL) for any configuration that requests "Sensitive" capabilities (shell access, network exposure).
* **Non-Goals:**
    * Sandboxing the actual execution of tools (that is the job of the tool provider or Docker).
    * Validating the *logic* of the tools, only the *configuration* of the gateway.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer cloning an open-source repo.
* **Primary Goal:** Safely use MCP Any with project-specific tools without risking RCE.
* **The Happy Path (Tasks):**
    1. User runs `mcpany serve` in a new repository.
    2. MCP Any detects `.mcpany/config.yaml`.
    3. The **Config Sandbox Validator** reads the file.
    4. It identifies a new "Hook" that runs `npm install`.
    5. The system flags this as "Sensitive" and pauses.
    6. UI displays a prompt: "This project wants to run a shell hook. Do you approve?"
    7. User approves, and the config is activated.

## 4. Design & Architecture
* **System Flow:**
    `Discovery -> Restricted YAML Parser -> Policy Evaluator (Rego) -> HITL Approval -> Active Config`
* **APIs / Interfaces:**
    * `internal/config/sandbox`: New package for isolated parsing.
    * `PolicyEngine.ValidateConfig(cfg)`: New method in the Policy Firewall.
* **Data Storage/State:**
    * "Pending" configs are stored in memory until approved.

## 5. Alternatives Considered
* **Ignore Project Configs:** Rejected because project-specific tools are a core part of the developer workflow.
* **Static Analysis Only:** Rejected because many malicious patterns are obfuscated and require runtime policy checks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The validator itself must be written in a memory-safe language (Go) and use a restricted YAML parser that prevents expansion attacks (Billion Laughs).
* **Observability:** All validation failures and approvals are logged in the **Immutable Execution Logs**.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
