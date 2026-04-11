# Design Doc: Ephemeral Mesh Identity (EMI) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Today's discovery of the "Shadow-Tunnel" exploit in OpenClaw reveals that session-bound hardware attestation tokens are vulnerable to intercept-and-replay attacks within P2P agent meshes. If a subagent intercepts a valid session ticket, it can reuse it to command remote tools on other nodes.

MCP Any needs to solve this by moving from persistent session identities to Ephemeral Mesh Identities (EMI). By rotating identity tokens for every single tool-call and binding them to a monotonic mission-root counter, we eliminate the window of opportunity for tunnel hijacking and ensure that every action is uniquely authorized.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested identity tokens that are valid for exactly one tool-call.
    * Bind every EMI token to a monotonic counter anchored in the mission root.
    * Provide sub-millisecond token minting to minimize coordination latency.
    * Neutralize the "Shadow-Tunnel" exploit pattern.
* **Non-Goals:**
    * Replacing long-lived user session tokens (EMI is for inter-agent/node coordination).
    * Managing the underlying P2P encryption (handled by the T2T Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Orchestrator
* **Primary Goal:** Securely invoke a filesystem tool on a remote laptop node from a primary desktop agent without risking tunnel hijacking.
* **The Happy Path (Tasks):**
    1. The Primary Agent requests an EMI token from the local MCP Any EMI Provider.
    2. The EMI Provider generates a token bound to the Hardware Inode and the next Monotonic Counter value.
    3. The Primary Agent sends the tool-call request + EMI token over the P2P tunnel.
    4. The Remote Node's AMT Broker validates the token against the monotonic sequence and the hardware root.
    5. The tool is executed, and the token is immediately marked as "Consumed" on the Remote Node.
    6. Any subsequent attempt to use the same token is rejected.

## 4. Design & Architecture
* **System Flow:**
    `[Primary Agent] -> [EMI Provider (Mint Token)] -> [P2P Tunnel] -> [Remote Node (Validate & Consume Token)] -> [Tool Execution]`
* **APIs / Interfaces:**
    * `POST /v1/identity/mint`: Request a single-use EMI token.
    * `Header: X-EMI-Token`: Carries the hardware-attested ephemeral identity.
* **Data Storage/State:**
    * Monotonic Counters are stored in the hardware-locked mission state (TPM/Secure Enclave).
    * Remote nodes maintain a Bloom Filter of recently consumed EMI tokens to prevent replay within the same counter window.

## 5. Alternatives Considered
* **Short-lived Session JWTs:** Rejected because they still have a valid window (e.g., 5-30 seconds) which is enough for machine-speed hijacking.
* **Synchronous Hardware Handshakes:** Rejected because the 100ms+ latency tax per call would cause "Cognitive Stall" in high-density swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are cryptographically bound to both the sender's hardware ID and the receiver's authorized mission fragment.
* **Observability:** Every EMI rotation event is logged in the Command Traceability Provider (CTP).

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
