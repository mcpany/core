# Design Doc: Hardware-Attested Action-Chain (HAAC) Validator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "hackerbot-claw" and similar autonomous exploits reveals a critical vulnerability in how AI agents manage sequential workflows. By poisoning environment initialization (e.g., `Go init()` functions) or build caches, attackers can hijack an agent's "vibe-coded" reasoning to execute malicious actions that appear valid in isolation but constitute a supply-chain attack when viewed as a chain.

MCP Any needs to solve this by providing a mechanism to notarize and validate the entire sequence of actions an agent takes, ensuring each step is a direct, untampered descendant of the user's verified mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographically signed "Chain of Custody" for agent tool calls.
    * Use TPM/Secure Enclave primitives to bind action tokens to hardware.
    * Enable real-time detection of "Action-Chain Drift" where an agent deviates from the authorized workflow manifest.
* **Non-Goals:**
    * Providing general-purpose build environment sandboxing (handled by external runtimes).
    * Validating the *logic* of the code generated (handled by APRIG).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Engineer managing autonomous CI/CD agents.
* **Primary Goal:** Ensure that a PR-fixing agent cannot be coerced into adding a malicious backdoor during its multi-step "Debug-Fix-Verify" loop.
* **The Happy Path (Tasks):**
    1. The user initiates a mission with a signed "Action Manifest."
    2. MCP Any issues a hardware-attested Root Action Token.
    3. The agent performs a tool call; MCP Any validates the token and appends a hash-chain fragment.
    4. The agent attempts a "Poisoned" initialization step not in the manifest.
    5. The HAAC Validator detects the signature mismatch and halts execution.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Signed Manifest| B(HAAC Notary)
        B -->|Attested Token| C[Agent Runtime]
        C -->|Tool Call + Token| D{HAAC Validator}
        D -->|Valid Chain| E[Target Tool]
        D -->|Invalid/Drift| F[Emergency Halt]
        E -->|Result + New Fragment| C
    ```
* **APIs / Interfaces:**
    * `POST /v1/action/notarize`: Issues the initial attested chain fragment.
    * `GRPC ValidateActionChain`: High-speed validation middleware for tool executors.
* **Data Storage/State:**
    * State is managed as a monotonic, hash-chained ledger in kernel-bound memory.

## 5. Alternatives Considered
* **Log-based Auditing:** Rejected due to "Reactive only" nature; cannot prevent the execution of the malicious step.
* **Static Sandbox Limits:** Rejected because dynamic workflows require flexibility that static rules cannot provide.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are hardware-bound and single-use per action step to prevent replay attacks.
* **Observability:** HAAC provides a "Reasoning Trace" that can be exported to the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
