# Design Doc: Project-Level Policy Engine Adapter
**Status:** Draft
**Created:** 2026-06-03

## 1. Context and Scope
With the release of Gemini CLI v0.30.0, the ecosystem is shifting toward repository-resident security policies. Managing security at a global level is becoming a bottleneck and often fails to address the unique tool and resource requirements of individual projects. MCP Any needs a standardized adapter to ingest and enforce these project-local policies (e.g., `mcp-policy.rego`) while ensuring they are hardware-attested and not subject to unauthorized overrides.

## 2. Goals & Non-Goals
* **Goals:**
    * Support project-resident security policies (e.g., `mcp-policy.rego`, `.mcp-policy.json`).
    * Implement a hardware-attested discovery mechanism for local policies.
    * Provide framework-neutral enforcement of project-level constraints (Gemini, OpenClaw, Claude Code).
    * Facilitate teammate-to-teammate policy synchronization in horizontal meshes.
* **Non-Goals:**
    * Overwriting global security policies (project policies are sub-scopes).
    * Providing a full Rego/CEL authoring environment (focus is on ingestion and enforcement).

## 3. Critical User Journey (CUJ)
* **User Persona:** Repository Maintainer / Swarm Orchestrator
* **Primary Goal:** Enforce a strict "No-Internet" policy for a specific project, regardless of which agent (Claude or OpenClaw) a contributor uses.
* **The Happy Path (Tasks):**
    1. Maintainer commits `mcp-policy.rego` to the repository root.
    2. Contributor opens the project with an MCP Any-connected agent.
    3. MCP Any performs a hardware-attested discovery of the local policy.
    4. MCP Any validates the policy signature and prompts the user for initial attestation.
    5. Once approved, all subsequent tool calls from any agent are validated against the project-local Rego rules.
    6. If a subagent attempts an unauthorized network call, the Policy Gate blocks it and triggers an alert.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Call] -> [Policy Engine Adapter] -> [Local Policy Cache (Verified)] -> [Enforcement Point]`
* **APIs / Interfaces:**
    * `policy.v1.IngestLocalPolicy(path, signature)`: Hardware-attested ingestion.
    * `policy.v1.ValidateAction(mission_id, tool_call)`: Multi-framework validation hook.
* **Data Storage/State:**
    * `policies.db`: SQLite storage for verified project policy hashes and attestation status.

## 5. Alternatives Considered
* **Framework-Specific Implementation:** Rejected because it would require redundant logic for Gemini, Claude Code, and OpenClaw.
* **Global Policy Expansion:** Rejected as it doesn't scale to thousands of heterogeneous repositories.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Local policies must be hardware-attested to prevent a malicious repository from silently downgrading security.
* **Observability:** Integration with the "Local Security Audit Dashboard" for policy violation tracking.

## 7. Evolutionary Changelog
* **2026-06-03:** Initial Document Creation.
