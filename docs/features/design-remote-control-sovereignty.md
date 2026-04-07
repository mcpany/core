# Design Doc: Remote-Control Sovereignty Broker
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
The introduction of "Remote Control" capabilities in major agent frameworks (e.g., Claude Code Dispatch/Remote Control) allows agents to be initiated in one context (e.g., a local terminal) and observed or steered from another (e.g., a remote browser or separate CLI). This "headless" model introduces a critical security gap: how to ensure the "observer" is authorized and how to securely hand off control without exposing secrets. The Remote-Control Sovereignty Broker provides the infrastructure to manage these handoffs using hardware-attested identity tokens.

## 2. Goals & Non-Goals
* **Goals:**
    * Mediate all remote "attach" requests to headless agent sessions.
    * Mandate hardware-attested handshakes for session takeover or observation.
    * Provide a cryptographically signed "Authority Chain" that persists across handoffs.
    * Enforce mission-root boundaries on remote steering inputs.
* **Non-Goals:**
    * Building a proprietary remote-access protocol (we wrap existing ones).
    * Providing the UI for steering (we provide the security backend).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed AI Engineer
* **Primary Goal:** Securely attach to a long-running agent session executing on a remote staging server.
* **The Happy Path (Tasks):**
    1. User initiates a Claude Code session on a remote server using the `Dispatch` mode.
    2. The agent registers its session with the MCP Any Remote-Control Sovereignty Broker.
    3. User opens their local terminal and attempts to `attach` to the session.
    4. MCP Any challenges the local user for a hardware-attested (TPM/Secure Enclave) identity token.
    5. User provides the token; MCP Any verifies it against the authorized mission-root initiators.
    6. MCP Any establishes a secure, encrypted bridge for the steering commands.
    7. User observes the agent's progress and injects a "Corrective Intent" which is validated by MCP Any's policy engine before being passed to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Remote Client` -> `Hardware Handshake` -> `Sovereignty Broker` -> `Headless Session`
* **APIs / Interfaces:**
    * `SovereigntyBroker`: `AttachSession(sessionID string, proof IdentityProof) (SessionBridge, error)`
    * `AuthorityManager`: `SignHandoff(targetIdentity string) (HandoffToken, error)`
* **Data Storage/State:**
    * Session metadata and authority chains are stored in a secure, ephemeral cache.
    * Mission-root constraints are persisted in the Shared KV Store.

## 5. Alternatives Considered
* **Password-based Remote Control**: Rejected due to vulnerability to phishing and brute-force.
* **Open Loopback Ports**: Rejected due to the "ClawJacked" (CVE-2026-25253) vulnerabilities.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Authority is Leased, Never Owned." Every steering command is re-validated against the active authority chain.
* **Observability:** All attach/detach events and steering inputs are logged in the "Remote Session Monitor" UI.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
