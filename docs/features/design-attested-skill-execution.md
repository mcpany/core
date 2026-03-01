# Design Doc: Cryptographic Skill Attestation
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The recent OpenClaw/ClawHub security crisis demonstrated that a public marketplace of agent skills is a prime target for supply chain poisoning. Adversaries can upload seemingly benign tools that contain infostealers or RCE vectors. MCP Any, as a universal gateway, must ensure that every tool it executes has a verifiable origin and hasn't been tampered with.

## 2. Goals & Non-Goals
* **Goals:**
    * Ensure every MCP server/skill executed by the gateway has a valid cryptographic signature.
    * Provide a mechanism for "Community Trust" and "Developer Attestation."
    * Block execution of unsigned or invalidly signed skills by default in production.
* **Non-Goals:**
    * Performing deep static analysis of the skill's source code (this is handled by the Runtime Behavioral Analysis feature).
    * Providing a full PKI infrastructure (will leverage existing standards like Sigstore/Cosign).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise AI Architect
* **Primary Goal:** Prevent unauthorized or malicious third-party skills from executing in the corporate environment.
* **The Happy Path (Tasks):**
    1. Architect configures MCP Any with a "Trusted Root" (e.g., a list of approved developer public keys).
    2. An agent attempts to call a tool from a new MCP server downloaded from a registry.
    3. MCP Any intercepts the call and checks for a `MANIFEST.sig` file associated with the server.
    4. MCP Any verifies the signature against the Trusted Root.
    5. The signature is valid; the tool call is permitted and logged.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request -> Gateway Interceptor -> Attestation Engine -> [Check Signature] -> Policy Engine -> Execution`
* **APIs / Interfaces:**
    * `POST /v1/attest/verify`: Internal endpoint to verify a server's bundle.
    * `GET /v1/attest/status`: Returns the attestation status of all registered MCP servers.
* **Data Storage/State:**
    * A local "Known-Good" database of hashes and signatures to prevent repeated remote verification.

## 5. Alternatives Considered
* **Manual Whitelisting:** Rejected as it doesn't scale with dynamic agent swarms and doesn't prevent tampering of already-whitelisted files.
* **Sandboxing Only:** Rejected because sandbox escapes (as seen in OpenClaw) are common; attestation provides a layer of defense *before* execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If a signature is invalid, the entire MCP server is quarantined. No "partial" trust for specific tools within a server.
* **Observability:** Every attestation failure triggers a high-severity alert in the audit log.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
