# Design Doc: Zero-Knowledge Discovery (ZKD) Bridge
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As agent swarms become more decentralized (A2A), "Shadow Mapping" has emerged as a significant threat. Malicious subagents or local scripts probe the gateway to map all available tools and capabilities, identifying weak points for exploitation. ZKD Bridge implements cryptographic masking to ensure that an agent's capabilities are completely invisible until a secure, mission-bound handshake is completed.

## 2. Goals & Non-Goals
* **Goals:**
    * Mask tool schemas and "Agent Cards" from unauthenticated discovery requests.
    * Require a mission-root-signed token to reveal specific capability sets.
    * Support Gemini CLI v0.42.0 ZKD protocol for cross-framework compatibility.
* **Non-Goals:**
    * It does not replace the A2A authentication layer; it enhances the *visibility* policy within it.
    * It does not manage tool execution permissions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent (e.g., a "Code Reviewer") from knowing that high-privilege tools (e.g., "Deployment Script") exist.
* **The Happy Path (Tasks):**
    1. A new agent joins the mesh and broadcasts a discovery request.
    2. MCP Any responds with a "Masked Capability Receipt" (containing no tool details).
    3. The orchestrator provides a hardware-attested mission-root token to the new agent.
    4. The agent presents this token to the ZKD Bridge.
    5. ZKD Bridge validates the token and unmasks only the toolset authorized for that specific mission branch.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Peer Agent] -- Discovery Request --> B[ZKD Bridge]
        B -- Masked Response --> A
        A -- Mission Token --> B
        B -- Validate TPM Signature --> C{Authorized?}
        C -- Yes --> D[Reveal Scoped Toolset]
        C -- No --> E[Maintain Masking]
    ```
* **APIs / Interfaces:**
    * `/v1/discovery/zkd-handshake`: New endpoint for processing mission-bound capability unmasking.
* **Data Storage/State:**
    * `capability_masks`: In-memory index mapping mission-roots to authorized tool subsets.

## 5. Alternatives Considered
* **Static Allow-Lists:** Rejected because they don't scale with dynamic swarm sizes and can be bypassed if the identity is spoofed.
* **Plain mTLS:** Rejected because it only proves *who* the agent is, not what *mission* they are currently on.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Even metadata (tool names) must be masked. A "Code Reviewer" shouldn't even know the *name* of a deployment tool.
* **Observability:** Track every "Unmasking Event" in the security audit log, including the mission-root lineage.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
