# Design Doc: IAM-Integrated Cloud-to-Local Bridge
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
With the rise of managed MCP services (e.g., Google's Managed MCP) and cloud-based agent sandboxes (e.g., Anthropic's), agents often run in environments with specific IAM identities. However, when these cloud agents need to delegate tasks to local subagents or tools via MCP Any, the security context (IAM role, permissions) is often lost. This feature enables MCP Any to act as a "Credential Reflector," allowing local subagents to securely inherit and utilize the parent agent's cloud identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Securely reflect cloud IAM identities (GCP Service Accounts, Anthropic tokens) to local tool executions.
    * Provide a standardized way for subagents to request "Identity Handoff."
    * Ensure Zero-Trust isolation between different cloud identities.
* **Non-Goals:**
    * Storing raw cloud credentials permanently on the local machine (ephemeral reflection only).
    * Managing the cloud IAM policies themselves (this is done in the cloud console).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cloud-Native Agent Developer
* **Primary Goal:** A parent agent in GCP Cloud Run calls a local tool in MCP Any and expects the tool to have the same BigQuery access.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a tool call to MCP Any, including a signed OIDC token or Attestation token.
    2. MCP Any validates the token and maps it to a "Reflected Identity."
    3. MCP Any executes the local tool in a subprocess or container where the environment is injected with ephemeral credentials derived from the Reflected Identity.
    4. The local tool successfully calls BigQuery using the parent agent's permissions.

## 4. Design & Architecture
* **System Flow:**
    `Cloud Agent -> MCP Any (Gateway) -> Identity Mapper -> Ephemeral Credential Provider -> Local Tool`
* **APIs / Interfaces:**
    * `X-MCP-Identity-Token`: Header for passing the cloud identity token.
    * `identity/reflect`: Internal endpoint for mapping tokens to local environment variables or volumes.
* **Data Storage/State:**
    * Reflected identities are stored in memory and bound to the session life-cycle.

## 5. Alternatives Considered
* **Static Credential Sharing**: Manually copying `.json` keys to the local machine. *Rejected* as insecure and hard to rotate.
* **Identity Tunneling (SSH/VPN)**: Creating a network tunnel for every agent. *Rejected* as too heavy and not scalable for ephemeral agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Identity tokens must be validated against the issuer (GCP/Anthropic). Ephemeral credentials must have a short TTL.
* **Observability:** Audit logs must record "Identity Reflected: [Principal] for Tool [ToolName]."

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
