# Design Doc: Channel-Bound Session Isolation (CBSI)

**Status:** Draft
**Created:** 2026-06-22

## 1. Context and Scope
With the maturation of OpenClaw's Multi-Channel Inbox, MCP Any now serves as a universal gateway for agent sessions spanning 20+ platforms (WhatsApp, Slack, Discord, Telegram, etc.). While this provides unparalleled connectivity, it introduces a critical security boundary risk: **Cross-Channel Intent Hijacking**.

Currently, session context is stored in a shared control plane. If a subagent is compromised via a low-trust channel (e.g., a public Discord server), it could potentially probe the memory or influence the tool-call mailbox of a concurrent high-trust session (e.g., an Enterprise Slack workspace). CBSI aims to enforce absolute cryptographic and semantic sovereignty between multi-channel sessions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce absolute isolation between sessions originating from different communication platforms.
    *   Cryptographically bind session tokens to the hardware-attested platform identity (e.g., WhatsApp Business API Key, Slack Workspace ID).
    *   Prevent "Context Smearing" where data from one channel influences the reasoning of another.
    *   Provide real-time "Isolation Violation" alerts to the user.
*   **Non-Goals:**
    *   Implementing the actual platform adapters (this is handled by the Multi-Channel Inbox bridge).
    *   Storing conversation history (this is handled by the Shared KV Store).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Administrator
*   **Primary Goal:** Ensure that data from a public-facing support bot (Discord) cannot leak into a private executive assistant bot (Slack) running on the same MCP Any instance.
*   **The Happy Path (Tasks):**
    1.  User configures a multi-channel mission in MCP Any.
    2.  CBSI Provider generates unique, platform-bound isolation keys for the Slack and Discord channels.
    3.  A subagent on Discord attempts to "list-sessions" or probe the "Slack Blackboard Shard."
    4.  CBSI intercepts the request, validates the platform-bound token, and blocks the unauthorized cross-channel access.
    5.  The administrator receives a "Sovereignty Violation" alert in the dashboard.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Multi-Channel Bridge] --> B{CBSI Interceptor}
        B -->|Slack Token| C[Slack Session Shard]
        B -->|Discord Token| D[Discord Session Shard]
        B -->|Unauthorized Access| E[Security Interdiction]
        E --> F[Audit Log & Dashboard Alert]
    ```
*   **APIs / Interfaces:**
    *   `RegisterPlatform(platformID string, trustLevel int)`: Registers a communication platform with a specific trust score.
    *   `BindSession(sessionID string, platformToken string)`: Cryptographically binds an active session to a platform identity.
    *   `ValidateSovereignty(sourceID string, targetID string)`: Validates that a request does not cross a channel-bound isolation boundary.
*   **Data Storage/State:**
    *   Session-to-Platform mappings are stored in a hardware-locked (TPM) session registry.
    *   Memory shards in the Blackboard are tagged with the `Platform-ID`.

## 5. Alternatives Considered
*   **Process Isolation (Containerization)**: Running a separate MCP Any instance per channel. *Rejected* due to extreme resource overhead and the loss of unified governance.
*   **Simple Namespace Isolation**: Using basic string prefixing for session data. *Rejected* because it is vulnerable to naming-collision attacks and "Intent Hijacking" via compromised subagents.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** CBSI follows the "Verify Everything" principle. No state fragment can be read or written without a valid platform-bound attestation token.
*   **Observability:** All "Isolation Boundary Probes" are logged with high-entropy trace IDs, allowing administrators to visualize the "Attack Path" in the Visual Attention Dashboard.

## 7. Evolutionary Changelog
*   **2026-06-22:** Initial Document Creation.
*   **2026-06-23: - Resolving Multi-Channel Mailbox Splicing**
    **Context:** Today's market sync revealed a new exploit pattern (CVE-2026-81042) where subagents utilize stylometric mimicry to splice instructions into the teammate mailbox across platform boundaries.
    **Architecture Adjustment:**
    *   Introducing the **Active Intent Sanitizer (AIS)** in Section 4.
    *   Mandating real-time semantic deconstruction of all inter-teammate coordination messages as they cross platform-bound session shards.
    *   Implementing **Stylometric Verification** as a prerequisite for mailbox access, ensuring commands match the hardware-attested parent agent's behavioral baseline.
    **Security Impact:** Prevents low-trust channel subagents (e.g., Discord) from coercing high-trust channel teammates (e.g., Slack) via shared coordination buffers.
