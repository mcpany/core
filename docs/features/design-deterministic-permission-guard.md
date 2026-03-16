# Design Doc: Deterministic Permission Guard (DPG)
**Status:** Draft
**Created:** 2026-05-09

## 1. Context and Scope
Existing AI agent frameworks often rely on "Instruction-Based" permissions (e.g., "Don't read /etc/shadow"). However, Bug #8961 in Claude Code and other "Instruction-Bypass" exploits show that agents can be coerced or hallucinate their way around these rules. The DPG is a kernel-level middleware that enforces project-local "Deny" rules independently of the agent's internal state.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce "Hard Deny" rules for project-local configurations.
    *   Integrate with Continuous CPCP for real-time validation.
    *   Provide OS-level file handle and network socket gating.
*   **Non-Goals:**
    *   Dynamic "Allow" rule generation (focus is on immutable safety boundaries).
    *   Replacement for host-level OS permissions (it's a middleware layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Prevent an autonomous subagent from ever reading `.env` files, even if its parent agent tells it to do so.
*   **The Happy Path (Tasks):**
    1.  The user defines a "Deny" rule for `**/*.env` in the project configuration.
    2.  The MCP Any gateway ingests and locks this policy via CPCP.
    3.  A subagent attempts to call `fs_read` on a `.env` file.
    4.  The DPG intercepts the call at the gateway level and returns an `ACCESS_DENIED` error before the tool is even invoked.

## 4. Design & Architecture
*   **System Flow:**
    `Tool Call` -> `DPG Policy Engine` -> `CPCP Validator` -> `Upstream MCP Server`
*   **APIs / Interfaces:**
    *   Integrated into the existing `Policy Firewall` middleware.
*   **Data Storage/State:**
    Policies are loaded from project-local `.mcpany/policies.json` and attested via TPM.

## 5. Alternatives Considered
*   **Agent-Side Filtering:** Rejected because the agent is part of the attack surface.
*   **Path-Only Validation:** Rejected due to symlink/hardlink escape vulnerabilities.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Policies themselves must be signed and hardware-attested.
*   **Observability:** Every DPG block is logged with high priority for the user.

## 7. Evolutionary Changelog
*   **2026-05-09:** Initial Document Creation.
