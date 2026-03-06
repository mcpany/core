# Design Doc: Agent Identity Management (Agent IAM)

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As agentic systems evolve into complex swarms, the current practice of using shared API keys or instance-level credentials has created an "Identity Crisis." Individual agents often operate without distinct identities, making it impossible to enforce the principle of least privilege or maintain accurate audit trails. Agent IAM aims to provide a robust framework for assigning, verifying, and managing cryptographic identities for every actor in the agentic ecosystem.

## 2. Goals & Non-Goals
* **Goals:**
    * Assign unique, verifiable cryptographic identities (e.g., DID or X.509) to individual agents.
    * Implement a "Least Privilege" access model where tools are bound to specific agent identities.
    * Enable cross-framework identity verification (e.g., OpenClaw agents interacting with AutoGen agents).
    * Provide a centralized "Identity Registry" within MCP Any for managing agent lifetimes.
* **Non-Goals:**
    * Replacing existing human IAM systems (e.g., Okta, Auth0).
    * Managing the internal logic of the agents themselves.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect
* **Primary Goal:** Ensure that a "Junior Researcher" agent cannot access the "Production Database" tool, even if it belongs to the same swarm as a "Senior Analyst" agent.
* **The Happy Path (Tasks):**
    1. Architect defines an Identity Policy in MCP Any: `AgentRole: JuniorResearcher -> Deny: tool:prod_db_access`.
    2. Swarm Orchestrator registers a new agent instance with MCP Any, receiving a unique `AgentID` and cryptographic key.
    3. The Junior Researcher agent attempts to call the `prod_db_access` tool.
    4. MCP Any verifies the `AgentID` signature and checks the policy.
    5. MCP Any rejects the call and logs a "Policy Violation" against the specific `AgentID`.

## 4. Design & Architecture
* **System Flow:**
    - **Identity Provisioning**: Agents are provisioned with short-lived, signed tokens during initialization.
    - **Signature Verification**: Every tool call must include an `X-Agent-Identity` header containing a signed payload.
    - **Policy Enforcement**: The `Policy Firewall` is updated to support `actor`-based rules, mapping `AgentID` to specific capabilities.
* **APIs / Interfaces:**
    - `POST /iam/agent/register`: Register a new agent and receive identity credentials.
    - `GET /iam/agent/:id/status`: Check the health and authorization status of an agent.
    - Extension to MCP Tool Call: `{"method": "tools/call", "params": {"agent_sig": "...", ...}}`.
* **Data Storage/State:**
    - Identities and their public keys are stored in the `Shared KV Store` (SQLite), with sensitive private keys managed by the orchestrator.

## 5. Alternatives Considered
* **Instance-Level Credentials**: Rejected because it doesn't provide granular control over individual agents within a swarm.
* **IP-Based Identification**: Rejected due to the dynamic nature of containerized/serverless agent deployments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero Trust component. It ensures that "Who" is calling the tool is as important as "What" is being called.
* **Observability:** All audit logs will now be enriched with `AgentID`, providing a clear map of which agent performed which action.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
