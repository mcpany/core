# Design Doc: Universal Commerce Adapter (UCA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents are increasingly being used for autonomous procurement and commerce, as highlighted by the new Universal Commerce Protocol (UCP). However, giving agents direct access to financial tools without strict governance poses catastrophic risks of unauthorized spending and token-draining attacks.
MCP Any needs to provide a secure, transactional gateway that bridges UCP-compliant tools with mission-root authorized financial boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Transaction-Bound Gating" for financial tool calls.
    * Mandate multi-agent commerce quorums for high-value transactions.
    * Provide hardware-attested audit trails for all UCP interactions.
* **Non-Goals:**
    * Directly managing user wallet private keys.
    * Acting as a payment processor (only acts as the gateway/adapter).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Procurement Agent
* **Primary Goal:** Securely purchase a cloud server instance within a $50 budget without exposing raw credit card details.
* **The Happy Path (Tasks):**
    1. Agent receives mission-root authorization for "Cloud Infrastructure Procurement".
    2. Agent discovers a UCP-compliant tool for a specific cloud provider via MCP Any.
    3. Agent initiates a purchase request via the UCA.
    4. UCA identifies the tool as "Financial/Commerce" and triggers a "Transaction-Bound Gate".
    5. UCA requests a mission-token that explicitly authorizes a $50 spend.
    6. A secondary "Auditor Agent" verifies the tool call against the mission manifest and signs the transaction.
    7. UCA executes the tool call and returns the hardware-attested receipt.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [UCP Tool Call] -> [UCA Middleware] -> [Mission Root Validation] -> [Commerce Quorum] -> [Encrypted Tool Proxy] -> [Vendor API]`
* **APIs / Interfaces:**
    * `POST /uca/v1/authorize-transaction`: Validates mission-root spend limits.
    * `GET /uca/v1/receipt-attestation`: Returns TPM-signed proof of transaction integrity.
* **Data Storage/State:**
    * Short-lived "Transaction Leases" stored in memory, backed by the UEG Memory Broker for audit persistence.

## 5. Alternatives Considered
* **Direct Tool Gating**: Rejected because financial tools require semantic understanding of "spend," not just binary allow/deny permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Multi-agent quorums prevent a single compromised specialist from draining budgets.
* **Observability:** Real-time spend tracking on the Visual Attention Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
