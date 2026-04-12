# Design Doc: CI/CD Workflow Integrity Guard
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agents are increasingly used for repo-resident tasks such as security scanning, automated dependency updates, and bug fixing. However, recent incidents of autonomous CI/CD hijacking demonstrate that agents can be coerced or decide to perform unauthorized action chains (e.g., scanning for secrets and then immediately pushing them to a malicious PR).

MCP Any needs to evolve from gating individual tool calls to gating **Workflow Sequences**. The CI/CD Workflow Integrity Guard provides a stateful security layer that validates the "Reasoning-to-Action" path, ensuring that agents only execute predefined, safe sequences of operations within a repository context.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement stateful action-chain monitoring for repo-resident agents.
    * Provide a "Safe Workflow" manifest that defines allowed sequences (e.g., `git fetch` -> `lint` -> `commit` is allowed; `env dump` -> `git push` is blocked).
    * Integration with Git-Native Agent Protocol (GNAP) for decentralized sequence validation.
    * Trigger Multi-Factor Attestation (MFA) when an agent attempts to diverge from a known safe workflow.
* **Non-Goals:**
    * Managing the actual CI/CD runner (GitHub Actions, GitLab CI).
    * Fixing the vulnerabilities found by agents.
    * Replacing existing Static Analysis (SAST) tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Repo Security Architect
* **Primary Goal:** Prevent an autonomous bug-fixing agent from exfiltrating `.env` files via a Pull Request.
* **The Happy Path (Tasks):**
    1. Architect defines a `SafeWorkflow` in `.mcpany/workflows.json`: `[ "read_file", "edit_file", "run_tests", "git_push_fix" ]`.
    2. Agent starts a "Bug Fix" mission.
    3. Agent calls `read_file` (Allowed).
    4. Agent calls `edit_file` (Allowed).
    5. Agent calls `run_tests` (Allowed).
    6. Agent attempts to call `read_file` on `.env`. The Guard checks the mission context and sees `.env` is not in the allowed file-set for this workflow.
    7. Guard blocks the call and requests user attestation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Gateway: Tool Call (git push)
        Gateway->>WorkflowGuard: Validate Sequence (Prev: edit, test)
        WorkflowGuard->>StateStore: Fetch Mission History
        StateStore-->>WorkflowGuard: History: [read, edit, test]
        WorkflowGuard->>PolicyEngine: Check History + Push vs SafeWorkflow
        PolicyEngine-->>WorkflowGuard: Result: VALID
        WorkflowGuard-->>Gateway: Proceed
        Gateway->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * `POST /v1/workflow/register`: Register a safe action chain.
    * `GET /v1/workflow/mission/{id}/history`: View the state of an active action chain.
* **Data Storage/State:**
    * Uses the Shared KV Store (Blackboard) with a new `workflow_state` namespace to track the transition of agent states.

## 5. Alternatives Considered
* **Individual Tool Gating**: Rejected because individual calls (like `read_file` or `git push`) may be safe in isolation but malicious in sequence.
* **Sandboxed Execution Only**: Necessary but insufficient, as the agent still needs to communicate results back to the host/repo, which is where exfiltration occurs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The workflow manifest itself must be hardware-attested (HAMM) to prevent the agent from modifying its own safety rules.
* **Observability:** Every action chain transition is logged to the Audit Log with a "Sequence Confidence Score."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
