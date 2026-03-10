# Design Doc: Cryptographic Agent Attestation (CAA)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
In multi-agent swarms, the trust boundary is traditionally the network connection or a shared API key. However, as agents become more autonomous and specialized (e.g., OpenClaw subagents), this model is insufficient. A "Researcher" subagent should not be able to masquerade as a "Finance" agent to access sensitive tools. Today's research (CVE-2026-AGENT-01) highlights that identity spoofing is a real threat. MCP Any must implement a Cryptographic Agent Attestation (CAA) layer where every interaction (tool calls, state changes, handoffs) is signed by the agent's unique identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Establish a "Registry of Trusted Agents" within MCP Any.
    * Require every tool call to be signed using an agent-specific private key (Ed25519).
    * Provide a mechanism for parent agents to "attest" for their children.
    * Integrate with the Policy Firewall to enable "Identity-Bound" tool permissions.
* **Non-Goals:**
    * Implementing a global PKI (Public Key Infrastructure).
    * Managing the agents' private keys (they must be stored securely by the agent runtime/orchestrator).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator.
* **Primary Goal:** Prevent a rogue "Plugin" agent from calling the "System Shell" tool by masquerading as the "Admin" agent.
* **The Happy Path (Tasks):**
    1. Orchestrator registers "AdminAgent" with its public key in MCP Any.
    2. AdminAgent sends a tool call signed with its private key.
    3. MCP Any verifies the signature against the Registry.
    4. Policy Firewall checks if "AdminAgent" has the `shell:exec` capability.
    5. Tool call is executed.
* **The Failure Path:**
    1. "PluginAgent" (unregistered) attempts to call `shell:exec` using a spoofed header `X-Agent-ID: AdminAgent`.
    2. MCP Any detects the missing or invalid signature.
    3. Request is rejected with "401 Unauthorized: CAA Verification Failed."

## 4. Design & Architecture
* **System Flow:**
    `Agent` --(Signed MCP Request)--> `CAA Middleware (MCP Any)` --(Verified Identity)--> `Policy Firewall` --> `Tool Execution`
    1. **Signature Format**: Requests use a custom JSON-RPC wrapper or an HTTP header `X-MCP-Signature` containing the Ed25519 signature of the payload.
    2. **Registry**: A local SQLite-backed registry stores `AgentID -> PublicKey` mappings.
    3. **Subagent Inheritance**: Parent agents can issue "Delegate Tokens" to subagents, which are valid for a specific duration and scope.
* **APIs / Interfaces:**
    * `POST /caa/register`: Register a new agent identity.
    * `POST /caa/verify`: Manually verify a signature (for debugging).
* **Data Storage/State:**
    * `caa_registry.db`: Table `(agent_id, public_key, parent_id, trust_level, created_at)`.

## 5. Alternatives Considered
* **Shared Secret (HMAC)**: Rejected because it doesn't support non-repudiation and is harder to manage in dynamic swarms.
* **OAuth2/mTLS**: Rejected as too heavy for local agent-to-gateway communication.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Fundamental to the Zero Trust goal. Identity is the root of all permissions.
* **Observability**: UI will display "Verified" badges next to tool calls in the Activity Feed.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
