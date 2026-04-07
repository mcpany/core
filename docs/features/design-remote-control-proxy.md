# Design Doc: Remote Control Security Proxy (RCSP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the maturation of Claude Code's "Remote Control" and "Dispatch" features, agents are increasingly running in headless server environments while being steered from local terminals. This "Remote Steering" introduces a critical security gap: if the communication channel between the steerer and the agent is compromised, an attacker can inject malicious intents or tools. RCSP acts as the authoritative gatekeeper for all remote steering signals.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate hardware-attested identity signatures for all remote session "join" or "steer" requests.
    * Enforce origin-locking for remote control signals to prevent cross-environment hijacking.
    * Provide real-time "Steering Logs" for auditing which environment is currently controlling an agent.
* **Non-Goals:**
    * Providing the underlying transport (e.g., WebSockets); it proxies and validates the traffic on that transport.
    * Managing the LLM reasoning process.

## 3. Critical User Journey (CUJ)
* **User Persona:** Remote DevOps Engineer
* **Primary Goal:** Securely steer a background "Dispatch" agent running on a production server from a local laptop.
* **The Happy Path (Tasks):**
    1. Engineer initiates a `remote steer` command from their laptop.
    2. The local MCP Any client generates a TPM-signed "Steering Token" bound to the engineer's identity and the agent's session ID.
    3. The RCSP on the server intercepts the connection request and validates the Steering Token.
    4. RCSP verifies that the laptop's origin is allow-listed for this specific mission-root.
    5. Once verified, RCSP allows the laptop to inject new intents into the agent's reasoning loop.
    6. All commands sent via the remote link are logged and attributed to the engineer's hardware identity.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Remote Steerer (Laptop)] -->|TPM-Signed Signal| B[RCSP Gateway]
        B -->|Validate Identity| C[FSI Provider]
        B -->|Verify Mission| D[HAMM Provider]
        B -->|Allowed| E[Headless Agent]
        E -->|Tool Call| F[MCP Any Server]
    ```
* **APIs / Interfaces:**
    * `rcsp.ValidateSteering(signal, hardwareProof) -> boolean`: Core validation logic.
    * `rcsp.AuthorizeRemoteJoin(sessionID, identity) -> SteeringTicket`: Issues a time-bound ticket for remote control.
* **Data Storage/State:**
    * **Steering Session Registry:** In-memory map of active remote controllers and their hardware IDs.

## 5. Alternatives Considered
* **Shared Secret/API Key for Remote Control:** Rejected because keys can be leaked. Hardware-attestation (TPM) ensures that only the specific device can steer the agent.
* **Standard SSH Tunneling:** Rejected as it lacks "Intent Awareness" and cannot granularly restrict what the steerer can do within the agentic context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enforces per-signal attestation. Even if a tunnel is established, every intent change must be signed.
* **Observability:** Integrated with the "Global Agent Activity Map" in the UI, showing remote steering origins.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
