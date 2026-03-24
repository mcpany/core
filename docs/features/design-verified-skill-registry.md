# Design Doc: Verified Skill Registry
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
The "ClawHavoc" crisis demonstrated that open-source agent marketplaces are vulnerable to malicious skill injection. Current models lack a centralized, security-focused verification layer. MCP Any needs to provide a `Verified Skill Registry` that acts as a secure "App Store" for AI agents, ensuring that every skill is analyzed, sandboxed, and signed before it can be used in a production environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized registry for MCP skills and servers.
    * Provide automated behavioral profiling (e.g., tracking network/file system access in a sandbox).
    * Require cryptographic signatures for all "Trusted" tier skills.
    * Integrate with the Policy Firewall to enforce "Allow-List Only" skill installation.
* **Non-Goals:**
    * Vetting the "quality" or "accuracy" of the skill's LLM outputs.
    * Replacing existing marketplaces like ClawHub (but acting as a security filter for them).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Admin managing a fleet of AI agents.
* **Primary Goal:** Ensure that only verified, safe skills are installed by developers.
* **The Happy Path (Tasks):**
    1. Admin configures MCP Any to "Verified Only" mode.
    2. A developer attempts to install a new skill from a public repo.
    3. MCP Any intercepts the installation and checks the `Verified Skill Registry`.
    4. The skill is found in the registry with a "Trusted" signature and a clean behavioral report.
    5. MCP Any allows the installation to proceed.
    6. If the skill is NOT verified, MCP Any blocks it and provides a link to request verification.

## 4. Design & Architecture
* **System Flow:**
    `Skill Installation Request` -> `Registry Checker` -> `Behavioral Analysis Sandbox` -> `Signature Validator` -> `Installation Approval`
    1. **Registry Service**: A metadata store for skill hashes, signatures, and analysis reports.
    2. **Analysis Engine**: A detached runner that executes the skill in a restricted environment to monitor its "side effects."
    3. **Policy Integration**: The existing `Policy Firewall` uses the registry status as a condition for tool execution.
* **APIs / Interfaces:**
    * `GET /v1/registry/verify/:skill_id`: Check the verification status of a skill.
    * `POST /v1/registry/analyze`: Submit a skill for behavioral profiling.
* **Data Storage/State:**
    * `registry.db`: Stores hashes of analyzed skills and their "Safety Score."

## 5. Alternatives Considered
* **Manual Code Review**: Too slow and doesn't scale with the volume of community skills.
* **Static Analysis Only**: Easily bypassed by obfuscated or dynamic code; behavioral profiling is more robust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Skills are quarantined until they pass the "Safe" threshold. High-risk permissions (e.g., `fs:write:*`) trigger a mandatory human review.
* **Observability**: The UI will provide a "Skill Safety Report" for every connected tool.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
* **2026-03-13:** Addressing the "ClawHavoc" crisis by adding details on automated behavioral profiling. Skills now undergo a mandatory "Isolation Test" where they are executed with a mocked filesystem and network to verify their declared permissions.
* **2026-03-17: Behavioral Profiling & Burn-In Periods**
    * **Context:** "ClawHavoc" malicious skills are using "Delayed Payloads" to bypass initial static analysis.
    * **Architecture Adjustment:** Skills now undergo a "Burn-In" period in an isolated sandbox. Their network and filesystem access patterns are profiled against a known baseline for 24 hours (simulated) before being promoted to "Trusted."
    * **Security Impact:** Detects malicious exfiltration attempts that only trigger after a skill has been "vetted" by basic static checks.

### Update: 2026-04-07 - Collective Skill Defense & Continuous Attestation
**Context:** The final post-mortem of the "ClawHavoc" crisis reveals that 12% of the marketplace was compromised, including skills that had previously passed initial "Burn-In" tests by mimicking legitimate behavior.
**Architecture Adjustment:**
* **Continuous Behavioral Attestation**: Shifting from one-time "Burn-In" to real-time, ongoing behavioral monitoring. If a skill's execution patterns diverge from its attested baseline (e.g., a "wallet tracker" starts accessing `.ssh` directories), its signature is immediately revoked.
* **Federated Reputation Quorum**: Integrating consensus-based safety signals. A skill's "Trusted" status is now a function of both local analysis and a quorum of remote MCP Any security nodes.
**Security Impact:** Provides a dynamic, adaptive defense against "Delayed Payload" and "Mimicry" attacks, ensuring that a skill's safety is verified for its entire lifecycle, not just at installation.

### Update: 2026-06-27 - Multi-Signature Skill Attestation (MSSA)
**Context:** The "ClawHub" compromise revealed that even "vetted" skills can be weaponized if the registry's single point of trust is breached.
**Architecture Adjustment:**
* **Multi-Signature Requirement**: Transitioning from single-provider signing to MSSA. Dynamic skill grafting now requires cryptographically bound approval tokens from both the agent framework and a verified third-party security auditor.
* **Auditor Sidecars**: Introducing "Auditor Sidecars" in the analysis engine that provide real-time, independent behavioral monitoring for high-risk tools.
**Security Impact:** Mitigates the risk of "Rug-Pull" supply chain attacks by ensuring no single entity can authorize high-risk tool execution.
