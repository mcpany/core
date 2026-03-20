# Design Doc: Hardware-Attested Mission Manifest (HAMM) Provider
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
As AI agents move from single-task tools to autonomous swarm participants, the risk of "Privilege Escalation via Discovery" has become critical. If an agent can autonomously discover and execute any tool in a project environment, a single reasoning slip or injection can lead to unauthorized host-level actions. Existing "Allow-List" models are often static and difficult to manage in dynamic, multi-agent workflows.

The HAMM Provider solves this by introducing a cryptographically bound mission manifest. A "Mission Manifest" defines the exact set of tools, platform channels, and resource budgets authorized for a specific mission branch *before* execution begins. By anchoring this manifest to hardware attestation (TPM/Secure Enclave), MCP Any ensures that even if a subagent's reasoning is compromised, its capability to interact with the world is strictly governed by an immutable, signed boundary.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a TPM-signed "Mission Manifest" for every agent session.
    * Enforce cryptographic binding between a session token and its manifest.
    * Prevent tool execution for any tool not explicitly pre-declared in the HAMM.
    * Support "Manifest Inheritance" where subagents can only request a subset of their parent's manifest.
* **Non-Goals:**
    * Performing real-time behavioral analysis (handled by Ghost Shell).
    * Managing the generation of the mission intent itself (handled by the LLM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure a "Code Review Agent" can only use `git` and `filesystem:read` tools, even if the environment contains `filesystem:write` or `shell:exec`.
* **The Happy Path (Tasks):**
    1. The user defines a Mission Manifest (YAML) for the Code Review task.
    2. The user signs the manifest using their local TPM-bound identity.
    3. The agent initiates a session with MCP Any, providing the signed HAMM.
    4. MCP Any verifies the signature and binds the session to the manifest.
    5. The agent attempts to call `filesystem:write`.
    6. MCP Any blocks the call as it is not in the HAMM, returning a "Capability Not Authorized" error.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>HAMM Provider: Session Init (Signed Manifest)
        HAMM Provider->>TPM: Verify Signature
        TPM-->>HAMM Provider: Success
        HAMM Provider->>Session Store: Bind SessionID to ManifestHash
        Agent->>MCP Any: Tool Call (session_id, tool_name)
        MCP Any->>HAMM Provider: Authorize(SessionID, ToolName)
        HAMM Provider->>Manifest Store: Get Manifest by Hash
        Manifest Store-->>HAMM Provider: Manifest Content
        HAMM Provider-->>MCP Any: Allow/Deny
    ```
* **APIs / Interfaces:**
    * `POST /session/init`: Accepts a signed HAMM object.
    * `GET /manifest/verify`: Internal utility for middleware to check capability alignment.
* **Data Storage/State:**
    * HAMM manifests are stored in a content-addressable storage (CAS) layer.
    * Session-to-Manifest mappings are held in the ephemeral session store.

## 5. Alternatives Considered
* **Static Config Allow-Lists:** Rejected because they don't support the dynamic nature of agent spawning and task-specific delegation.
* **Purely Semantic Gating:** Rejected because it relies on the LLM's own reasoning, which is the very thing we are trying to protect against.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The manifest is immutable once signed. Any attempt to modify the HAMM in-flight invalidates the signature.
* **Observability:** Every "HAMM Deny" event is logged with full lineage metadata (parent agent, mission root).

## 7. Evolutionary Changelog
* **2026-03-20:** Initial Document Creation.
