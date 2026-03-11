# Design Doc: Deterministic Validation Middleware
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the rise of "Remediation Agents" like Claude Code and OpenClaw, AI is now frequently proposing and applying code changes. However, LLM reasoning is probabilistic and can introduce subtle bugs or security regressions. MCP Any needs a "Deterministic Validation Layer" that acts as a gatekeeper. This middleware intercepts tool calls that modify the system (e.g., `write_file`, `edit_file`) and mandates the successful execution of deterministic checks (tests, linters, security scanners) before the changes are committed to the host filesystem.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept "high-stakes" tool calls (filesystem modifications, system configuration changes).
    * Execute user-defined validation suites (e.g., `npm test`, `go test`, `eslint`) in a detached sandbox.
    * Block the commitment of changes if validation fails.
    * Support "Staged Execution" where changes are first applied to a temporary virtual workspace for validation.
* **Non-Goals:**
    * Automatically "fixing" the failed tests (this is the agent's job).
    * Validating non-system-modifying tools (e.g., `read_file`, `web_search`).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an autonomous agent to refactor a critical library.
* **Primary Goal:** Ensure that AI-generated refactors do not break existing functionality or violate linting rules.
* **The Happy Path (Tasks):**
    1. Agent calls `write_file` to update `auth.go`.
    2. MCP Any intercepts the call and diverts the write to a `Staged Workspace`.
    3. MCP Any triggers a `Validation Task`: runs `go test ./...` and `golangci-lint run`.
    4. Validation passes.
    5. MCP Any commits the change from the `Staged Workspace` to the host filesystem.
    6. Agent receives a success response.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any (Validation Middleware)` -> `Staged Workspace (Sandbox)` -> `Validation Engine` -> `Host Filesystem`
    1. **Interception**: Middleware wraps all "mutator" tools.
    2. **Staging**: Changes are written to an ephemeral overlay or copy of the project.
    3. **Execution**: The `Validation Engine` runs configured commands within the `Detached Sandbox`.
    4. **Enforcement**: If exit code != 0, the middleware returns an error to the agent with the validation logs.
* **APIs / Interfaces:**
    * `config.yaml`: `validation_rules` block defining which tools trigger which commands.
    * `Tool.Execute` wrapper: Logic for staging and validation.
* **Data Storage/State:**
    * `Staged Workspace`: Temporary directory managed by the middleware.
    * `validation_history.db`: Logs of all validation runs and results.

## 5. Alternatives Considered
* **CI/CD Pipelines**: Too slow for iterative agent development. Validation needs to happen *during* the tool call.
* **Agent-internal testing**: Agents often forget to run tests or ignore failures. Infrastructure-level enforcement is more reliable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Validation commands run in a `Detached Sandbox` with no network access (unless specified) and restricted resource limits to prevent "Validation-time RCE."
* **Observability**: Validation logs are streamed to the MCP Any UI for real-time monitoring and debugging.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
