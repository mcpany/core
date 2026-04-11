# Design Doc: Origin-Locked Local Trust Bridge
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
The disclosure of CVE-2026-25253 (OpenClaw token exfiltration) and CVE-2026-25593 (RCE via Gateway WebSocket) proves that "Implicit Local Trust" for loopback (127.0.0.1) traffic is a critical failure point. Malicious browser scripts or unauthorized local applications can currently bridge into the agent's control plane. The Origin-Locked Local Trust Bridge mandates cryptographically bound origin validation for all local listeners to ensure only authorized local applications can command the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement strict `Origin` and `Sec-Fetch-Site` header verification for all WebSocket and HTTP endpoints.
    * Introduce an "Origin Pairing" flow where the user must explicitly approve new local application origins.
    * Cryptographically bind session tokens to the initiating browser/CLI origin.
    * Provide sub-millisecond rejection of unauthorized loopback probes.
* **Non-Goals:**
    * Replacing standard TLS/mTLS for remote traffic (this focuses on the local trust gap).
    * Managing OS-level application permissions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local AI Developer.
* **Primary Goal:** Ensure that a malicious website open in their browser cannot command their local MCP Any instance to execute shell commands.
* **The Happy Path (Tasks):**
    1. User opens a verified local UI (e.g., `localhost:3000`).
    2. The UI attempts to connect to the MCP Any Gateway.
    3. MCP Any detects a new origin and pauses the connection.
    4. A "Local Pairing Request" appears in the user's system tray/notification area.
    5. User clicks "Approve" and signs the pairing with a hardware key.
    6. MCP Any issues an Origin-Bound Token (OBT) and allows the connection.
    7. A malicious site (`attacker.com`) attempts to connect; MCP Any rejects it instantly because the Origin does not match the allow-list.

## 4. Design & Architecture
* **System Flow:**
    `Request` -> `Origin Extractor` -> `Allow-list Check` -> `Pairing Logic (if new)` -> `Token Binding` -> `Dispatch`
* **APIs / Interfaces:**
    * `OriginValidator`: `Validate(req Request) (bool, error)`
    * `PairingService`: `InitiatePairing(origin string) (ApprovalToken, error)`
* **Data Storage/State:**
    * `allowed_origins.json`: Signed list of approved local origins.
    * Origin-bound session store in memory/Redis.

## 5. Alternatives Considered
* **Listening only on UNIX Domain Sockets**: Rejected because many agent frameworks and browser-based tools require WebSockets/HTTP for compatibility.
* **Mandatory mTLS for localhost**: Rejected due to the extreme friction of managing local CA certs for every browser instance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Fundamental to closing the "Loopback Hole".
* **Observability**: All blocked origin attempts are logged with high severity and surfaced in the UI.

## 7. Evolutionary Changelog
* **2026-04-11:** Initial Document Creation.
