# Design Doc: Agentic Privilege Auditor (APA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents transition from sequential tasks to complex, autonomous missions, they often inherit expansive permissions from the enterprise environments they operate in (e.g., full access to SharePoint, S3 buckets, or local filesystems). This "Inherited Permission Sprawl" creates a critical security gap known as the **Agentic Insider Threat**. MCP Any needs to provide an `Agentic Privilege Auditor (APA)` that dynamically scopes an agent's active privileges to its specific, hardware-attested mission intent, ensuring the principle of least privilege is enforced in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a real-time auditor that maps inherited system permissions to the agent's verified "Mission Root."
    * Provide automated "Privilege Masking" for data sources and tools not explicitly required for the current task.
    * Require hardware-bound (TPM) attestation for any "Privilege Escalation" within a mission branch.
    * Integrate with the `Policy Firewall` to enforce mission-bound access control.
* **Non-Goals:**
    * Managing the underlying IAM policies of the cloud providers (e.g., AWS IAM, Azure AD).
    * Predicting the "best" permissions for a task (this is managed by the Mission Manifest).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect configuring a multi-agent swarm.
* **Primary Goal:** Prevent a specialist subagent (e.g., a "Documentation Analyzer") from accessing sensitive financial data it technically has access to via its parent container.
* **The Happy Path (Tasks):**
    1. The lead agent spawns a subagent with a signed mission fragment: "Analyze documentation in /docs/v2."
    2. The subagent attempts to list files in `/sensitive/finances`.
    3. The `Agentic Privilege Auditor (APA)` intercepts the tool call.
    4. APA checks the subagent's mission fragment and determines the requested path is outside the authorized scope.
    5. APA "masks" the directory, returning an empty list or a permission error, even though the host process technically has access.
    6. APA logs the attempt as a "Scope Boundary Violation" for the audit trail.

## 4. Design & Architecture
* **System Flow:**
    `Agent Tool Call` -> `Mission Context Provider` -> `Privilege Auditor (APA)` -> `Scope Validator` -> `Execution/Interdiction`
    1. **Mission Context Provider**: Retrieves the cryptographically signed intent and mission-root manifest for the active session.
    2. **Privilege Auditor (APA)**: The core engine that compares the requested resource/tool against the mission manifest and the inherited system-level permissions.
    3. **Privilege Masking Layer**: A middleware that interceptor standard I/O and API calls to return sanitized or filtered views of the environment.
* **APIs / Interfaces:**
    * `GET /v1/apa/effective_permissions/:session_id`: Returns the current effective (masked) permissions for an agent session.
    * `POST /v1/apa/validate_access`: Validates a specific resource access request against the mission root.
* **Data Storage/State:**
    * `mission_scopes.db`: Maps hardware-attested session IDs to their authorized mission manifests and parentage.

## 5. Alternatives Considered
* **Static Sandboxing (Docker/gVisor)**: Effective but too rigid for dynamic swarms where agents need flexible access to a subset of project-local files that change per mission.
* **Manual IAM Tagging**: Impossible to scale for ephemeral, autonomous subagent spawns that exist for only minutes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: APA assumes all inherited permissions are "over-provisioned" by default and requires explicit mission-alignment to unmask them.
* **Observability**: The UI will provide a "Privilege Violation Heatmap" showing where agents are attempting to exceed their mission boundaries.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
