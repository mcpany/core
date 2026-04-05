# Design Doc: Semantic Diff Attestation (SDA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agents are increasingly used to modify source code. Current security models rely on coarse-grained file access controls or manual PR review. However, as agent speed increases, manual review becomes a bottleneck, and coarse controls don't prevent an agent from "correctly" implementing a feature while "incorrectly" introducing a back door.

The Semantic Diff Attestation (SDA) Hub generates cryptographic proofs that a proposed code change (a `git diff`) aligns semantically with the authorized mission root and does not violate established security invariants.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate "Intent-Alignment Proofs" for every code-modifying tool call.
    * Detect "Instruction Smuggling" where a diff contains changes not requested in the task description.
    * Integrate with the Action-Chain Sovereignty Monitor (ACSM) to block non-attested commits.
* **Non-Goals:**
    * Full formal verification of all code (focus is on semantic intent alignment).
    * Replacing CI/CD security scanners (SDA is a pre-execution/pre-commit gate).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer
* **Primary Goal:** Allow an agent to fix 50 dependency vulnerabilities automatically with 100% confidence that the agent *only* updated the versions and didn't touch logic.
* **The Happy Path (Tasks):**
    1. The agent is assigned a task card: "Update dependency X to version Y."
    2. The agent uses `write_file` to update `package.json`.
    3. The SDA Hub intercepts the call, compares the diff with the task card using a "Small Auditor Model" (local).
    4. The Auditor confirms the diff only contains the version change.
    5. A TPM-signed "Intent-Alignment Proof" is attached to the tool call.
    6. The commit is allowed.

## 4. Design & Architecture
* **System Flow:**
    [Tool Call] -> [SDA Middleware] -> [Semantic Auditor] -> [Attestation Signer] -> [ACSM Enforcement]
* **APIs / Interfaces:**
    * `sda.AttestDiff(diff, intent_fragment)` -> Returns a signed proof or rejection.
* **Data Storage/State:**
    * Proofs are stored in the session-bound audit log.

## 5. Alternatives Considered
* **Regex-based Validation:** Rejected as it cannot handle the semantic complexity of code changes.
* **Mandatory Human Review:** Rejected as it doesn't scale with autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The auditor itself must run in a detached sandbox with no network access.
* **Observability:** SDA results are surfaced in the "Action-Chain" UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
