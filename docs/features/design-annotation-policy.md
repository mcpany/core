# Design Doc: Annotation-Driven Policy Engine
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As agents become more autonomous, simple name-based tool filtering (allow/deny `fs:read`) is no longer sufficient. Agents need to understand the *nature* of the tools they are using, and security teams need to enforce policies based on tool characteristics. Gemini CLI's recent update to support tool annotation matching highlights this shift. MCP Any must evolve its Policy Firewall to support rich, metadata-driven governance.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable policy enforcement based on arbitrary tool annotations (e.g., `risk_level`, `pii_access`, `billing_impact`).
    * Support "Intent-Aware" filtering by matching agent-provided intent with tool annotations.
    * Provide a standardized set of "Well-Known Annotations" for common security concerns.
* **Non-Goals:**
    * Hard-coding specific annotations into the core engine (must be extensible).
    * Automatically generating annotations for third-party tools (though we may provide a manual mapping layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer.
* **Primary Goal:** Prevent an autonomous agent from using any tool marked with `billing_impact: high` unless specifically authorized for a budget-approved session.
* **The Happy Path (Tasks):**
    1. Compliance Officer defines a policy: `deny tool_call if tool.annotations.billing_impact == "high" && session.budget_approved == false`.
    2. Agent attempts to call a `provision_gpu_cluster` tool.
    3. The tool definition in MCP Any includes the annotation `{ "billing_impact": "high" }`.
    4. The Policy Firewall intercepts the call, evaluates the Rego/CEL policy, and blocks the execution because the session lacks budget approval.
    5. Agent receives a structured error explaining the policy violation.

## 4. Design & Architecture
* **System Flow:**
    * **Metadata Enrichment**: During tool discovery, MCP Any merges upstream annotations with local overrides defined in `config.yaml`.
    * **Policy Evaluation**: The Policy Firewall (Rego/CEL) is provided with the full tool metadata object during every `tools/call` interception.
* **APIs / Interfaces:**
    * `config.yaml` extension:
      ```yaml
      services:
        my-tool:
          annotations:
            risk_level: "high"
            pii_access: "true"
      ```
* **Data Storage/State:**
    * Annotations are stored in the Tool Registry (SQLite).

## 5. Alternatives Considered
* **Namespace-Based Security**: Using tool prefixes (e.g., `prod.tool_name`) to imply risk. *Rejected* as it is brittle and limits expressiveness.
* **Separate Governance Service**: Outsourcing policy checks to an external service. *Rejected* for the "Universal Adapter" use case to keep local execution fast and self-contained.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Annotations themselves must be protected from tampering by the agent. Upstream annotations are treated as "untrusted" unless they come from an Attested Source.
* **Observability:** Audit logs will record which annotations triggered a policy decision.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation. Standardizing annotation-driven governance to align with Gemini CLI and enterprise security requirements.
