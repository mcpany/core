# Design Doc: Remote Channel Authenticator (RCA)
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
With the launch of "Claude Code Channels" and the rise of remote messaging-based agent control (Telegram, Discord), AI agents are increasingly operating in headless, multi-channel environments. Standard webhook-based integrations are vulnerable to identity spoofing and token hijacking, as they often rely on a single long-lived platform token.

The Remote Channel Authenticator (RCA) is required to act as the authoritative gateway for these platforms, enforcing hardware-attested, session-bound verification for every remote command to ensure that only the authorized user can command the local agent bus from a remote device.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure bridge between external messaging platforms and the local MCP Any bus.
    * Enforce hardware-attested authentication (TPM/Secure Enclave) for remote instructions.
    * Neutralize identity spoofing by cryptographically linking messaging user IDs to local verified sessions.
    * Implement "Mission-Bound Notification Sharding" to prevent context leakage.
* **Non-Goals:**
    * Replacing the messaging platform's own authentication (e.g., Telegram login).
    * Building a full-featured bot framework; RCA focuses on the security handshake.
    * Managing remote filesystem access directly; it proxies commands to existing tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile Developer / Sovereign Agent Orchestrator
* **Primary Goal:** Securely trigger a "git status" and "refactor" command on a local workstation via Telegram while on a commute.
* **The Happy Path (Tasks):**
    1. User pairs their Telegram account with their local MCP Any instance via a hardware-attested QR code or "Trust Link."
    2. User sends "/status" to the agent's Telegram bot.
    3. RCA receives the webhook and identifies the user's remote ID.
    4. RCA issues an "Auth Challenge" to the user's local machine (if active) or requires a pre-shared hardware-attested "Remote Lease."
    5. Once the session is verified, RCA passes the command to the Multi-Channel Intent Validator (MCIV).
    6. MCIV verifies the intent against the active mission root.
    7. The command is executed, and the result is returned through a mission-bound notification shard to Telegram.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[User Mobile / Telegram] -->|Remote Command| B[Messaging Webhook]
        B --> C[Remote Channel Authenticator]
        C -->|Verification| D[Local Hardware Enclave / TPM]
        C -->|Validated Instruction| E[Multi-Channel Intent Validator]
        E -->|Mission Bound| F[Agent Bus / MCP Tools]
        F --> G[Notification Shard]
        G -->|Filtered Result| A
    ```
* **APIs / Interfaces:**
    * `rca.RegisterChannel(platform, platformUserID, trustToken)`: Pairs a remote ID with a local hardware session.
    * `rca.ValidateInbound(webhookPayload) -> AuthenticatedInstruction`: Verifies the cryptographic integrity of the inbound message.
    * `rca.IssueRemoteLease(duration, missionScope)`: Generates a time-bound, hardware-locked token for remote control.
* **Data Storage/State:**
    * **Channel Mapping Registry:** Secure storage linking remote platform IDs to local session public keys.
    * **Lease Store:** Tracking active remote-control leases and their associated mission roots.

## 5. Alternatives Considered
* **Plain Webhook with Static Secret:** Rejected due to the risk of "Mirroring" attacks and token theft. Static secrets provide no hardware-bound identity assurance.
* **VPN/SSH Tunneling:** Rejected as too cumbersome for mobile messaging interactions and lacking "Intent-Aware" filtering.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All inbound webhooks must match a "Webhook Integrity Attestation" proof. No command is executed without explicit alignment with a local mission root.
* **Observability:** Remote interactions are logged in the "Local Security Violation Monitor" with distinct "Remote Channel" markers.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
