# Design Doc: Verifiable Ephemeral Filesystem (VEF) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Current local tool execution for agents often involves direct host filesystem access or coarse-grained Docker volume mounts. This creates a large blast radius for malicious or hallucinatory file operations. As agents move toward more autonomous coding and system administration tasks, we need a "Transaction-First" approach to the filesystem.

The **Verifiable Ephemeral Filesystem (VEF) Broker** provides a memory-mapped, cryptographically attested overlay for local execution. Agents see a realistic view of the project, but all writes are captured in an ephemeral buffer. These changes are only "committed" to the host after a multi-agent integrity quorum (including a security auditor agent) validates the diff against the mission-root security policy.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a Zero-Trust, memory-mapped filesystem view for agent execution.
    * Capture all filesystem mutations in a hardware-attested buffer.
    * Implement "Integrity Quorum" signing for diff commitments.
    * Enable instant rollback of speculative filesystem changes.
* **Non-Goals:**
    * Replacing long-term storage (e.g., databases).
    * Providing network-level isolation (handled by AMT/Isolated Pipes).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Agent Orchestrator
* **Primary Goal:** Allow an agent to refactor a codebase without risking host-level file corruption or unauthorized exfiltration via hidden files.
* **The Happy Path (Tasks):**
    1. The Agent requests access to the `/src` directory.
    2. The VEF Broker creates a memory-mapped ephemeral view of `/src`.
    3. The Agent performs `sed` and `rm` operations within its environment.
    4. The VEF Broker intercepts these calls and stores the mutations in a "Mission Buffer".
    5. The Agent signals task completion.
    6. The VEF Broker generates a "Filesystem Diff".
    7. A "Security Monitor" agent reviews the diff and provides a hardware-bound approval token.
    8. The VEF Broker commits the diff to the host filesystem.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        HFS[Host Filesystem] -- Read Only --> VEF[VEF Broker]
        VEF -- Mapped View --> AE[Agent Environment]
        AE -- Writes --> MB[Mission Buffer]
        MB -- Diff --> IQ[Integrity Quorum]
        IQ -- Success --> HFS
        IQ -- Failure --> MB
    ```
* **APIs / Interfaces:**
    * `POST /v1/vef/mount`: Initialize an ephemeral view for a session.
    * `GET /v1/vef/{id}/diff`: Retrieve the pending changes.
    * `POST /v1/vef/{id}/commit`: Finalize changes to the host.
* **Data Storage/State:**
    * Mission buffers utilize Linux `tmpfs` or `memfd` for high performance.
    * Attestation hashes are stored in the SRM (Signed Reasoning Monologue) provider.

## 5. Alternatives Considered
* **Docker Containerization:** Standard but lacks granular, per-file attestation and low-latency diff reconciliation for host-resident files.
* **FUSE Filesystems:** High flexibility but introduces performance overhead for high-frequency small writes typical of compilation tasks. VEF uses memory-mapping for lower latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The commit phase requires TPM-bound signatures. Unauthorized attempts to bypass the VEF Broker trigger immediate mission revocation.
* **Observability:** Diff-level logging integrated with the Action-Chain Sovereignty Monitor (ACSM).

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
