# Design Doc: Teammate Identity Attestation (TIA)
**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope
In high-density horizontal swarms, specialist agents communicate via inter-agent mailboxes. Market sync reports show a rising exploit pattern where compromised subagents impersonate more privileged teammates by spoofing mailbox headers. MCP Any must mandate hardware-attested identity for every coordination message to secure the horizontal mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-bound (TPM/Secure Enclave) identity signatures for every teammate-to-teammate (T2T) message.
    * Neutralize "Mailbox Injection" and teammate impersonation attacks.
    * Implement sub-millisecond attestation validation to prevent coordination latency.
* **Non-Goals:**
    * Replacing the mission-root's global authentication to the model provider.
    * Managing persistent identities across different mission roots.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Architect
* **Primary Goal:** Ensure that a specialist "Database Agent" cannot be commanded by an unauthorized "Frontend Agent" even if they share the same T2T bridge.
* **The Happy Path (Tasks):**
    1. Agent A (Frontend) sends a coordination request to Agent B (Database) via MCP Any.
    2. MCP Any intercepts the request and mandates a TIA signature from Agent A.
    3. Agent A signs the message with its hardware-attested Teammate Token.
    4. Agent B verifies the signature and mission-binding before executing the task.

## 4. Design & Architecture
* **System Flow:**
    * TIA is integrated as a mandatory middleware in the T2T Encryption Bridge.
    * Every `mailbox.send` call requires a `x-tia-attestation` header.
    * The SMI Relay acts as the local Certificate Authority, validating teammate tokens against the Mission Root.
* **APIs / Interfaces:**
    * `tia.v1.SignMessage(payload)`: Returns a hardware-bound signature for a mailbox message.
    * `tia.v1.VerifyIdentity(signature, mission_id)`: Validates the sender's identity within the mission scope.
* **Data Storage/State:**
    * The SMI Relay maintains an in-memory cache of hardware-verified teammate pubkeys.

## 5. Alternatives Considered
* **Shared Symmetric Keys:** Rejected because a single compromised teammate would expose the entire mesh.
* **Centralized Auth Server:** Rejected due to the "High-Frequency Tax" (latency) and the risk of a single point of failure in local-first swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** TIA enforces the "Mission-Bound Identity" (MBI) standard; tokens are cryptographically invalid outside their specific mission scope.
* **Observability:** Logs "Identity Anomalies" (spoofing attempts) to the Audit Log.

## 7. Evolutionary Changelog
* **2026-05-30:** Initial Document Creation.
