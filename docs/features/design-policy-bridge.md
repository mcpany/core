# Design Doc: Policy-as-Code Bridge (Seatbelt Profiles)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the release of Gemini CLI v0.30.0, the concept of "Seatbelt Profiles" has emerged as a standard for strict, user-defined behavioral policies for AI agents. As a universal gateway, MCP Any must bridge these ecosystem-specific policies into a unified framework. This prevents "security fragmentation" where an agent is secure in one CLI but vulnerable when accessed via another adapter.

## 2. Goals & Non-Goals
* **Goals:**
    * Support importing and exporting Gemini-style "Seatbelt" policies.
    * Provide a unified middleware that enforces these policies across all MCP transports (Stdio, HTTP, gRPC).
    * Enable "Policy Parity" where a single security profile can be applied to multiple agent frameworks (Claude, Gemini, OpenClaw).
* **Non-Goals:**
    * Creating a new policy language (we will leverage Rego/CEL and bridge to/from Seatbelt JSON/YAML).
    * Enforcing policies at the LLM provider level (focus is on the tool execution gateway).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious DevOps Engineer.
* **Primary Goal:** Apply a "Strict Read-Only" seatbelt profile to a swarm of 5 agents across different frameworks.
* **The Happy Path (Tasks):**
    1. User defines or imports a `seatbelt-readonly.yaml` file.
    2. MCP Any's Policy Bridge converts the Seatbelt profile into internal Rego hooks.
    3. The user attaches this profile to an agent session in MCP Any.
    4. An agent attempts to call `fs_write`; MCP Any intercepts the call, evaluates it against the "Seatbelt," and blocks the execution.
    5. The violation is logged with a reference to the specific Seatbelt rule.

## 4. Design & Architecture
* **System Flow:**
    - **Ingestion**: A parser that maps Seatbelt profile fields (e.g., `allowed_tools`, `restricted_paths`) to MCP Any's Policy Firewall schema.
    - **Enforcement**: The Policy Firewall middleware executes the bridged rules during the `pre-call` hook of the tool execution lifecycle.
    - **Export**: Ability to generate Seatbelt-compatible configurations from MCP Any's native policies.
* **APIs / Interfaces:**
    - `POST /api/v1/policies/import/seatbelt`: Accepts a Gemini-style policy file.
    - `GET /api/v1/policies/export/seatbelt`: Exports current policies in Seatbelt format.
* **Data Storage/State:** Policies are stored in the versioned `config.yaml` or a dedicated `policies/` directory.

## 5. Alternatives Considered
* **Manual Translation**: Forcing users to manually rewrite policies for MCP Any. *Rejected* as it leads to configuration drift and human error.
* **Direct Execution of Gemini Binaries**: Running the Gemini CLI's policy engine as a subprocess. *Rejected* due to performance overhead and lack of control over non-Gemini agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The bridge itself must be sandboxed to prevent a malicious policy from compromising the gateway (Policy Injection).
* **Observability:** Audit logs will include the original "Seatbelt" rule ID to assist security teams in cross-platform investigation.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
