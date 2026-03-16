# Design Doc: Kernel-Bound FD Persistence Middleware
**Status:** Draft
**Created:** 2026-05-04

## 1. Context and Scope
The "Recursive Symlink Tunnel" exploit and the persistence of "Symlink-to-Inode Racing" (SIR) demonstrate that path-based and even simple Inode-based validation can be bypassed if an attacker can swap files during the kernel's traversal or between the application's check and use phases. Kernel-Bound FD Persistence evolves our security model from pinning Inodes to pinning the actual File Descriptors (FD) and utilizing FD-passing. This ensures that once a project-local configuration (e.g., `.claude/settings.json`) is opened and validated by the gateway, that exact file remains the only source of truth for the duration of the agent's session, regardless of any symlink swaps on the host filesystem.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement FD-passing and pinning for all security-critical project-local configurations.
    * Ensure absolute immutability of the execution context from the moment of initial attestation.
    * Provide a hardware-bound guarantee (linked to TPM) that the pinned FD matches the attested state.
    * Support depth-aware validation for recursive symlink tunnels during the initial open phase.
* **Non-Goals:**
    * Implementing a custom kernel module.
    * Replacing OS-level file locking (flock/lockf).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely load a complex swarm configuration where some settings are dynamically generated but must be immutable once the mission starts.
* **The Happy Path (Tasks):**
    1. Agent requests to load a configuration file.
    2. MCP Any performs a depth-aware recursive resolution.
    3. Gateway opens the file and captures the FD and hardware Inode.
    4. FD is "Pinned" in the Kernel-Bound FD registry.
    5. All subsequent reads/writes use the pinned FD or a duplicated FD passed through the Shadow-FS.
    6. Any external attempt to swap the file path is ignored as the gateway remains bound to the open FD.

## 4. Design & Architecture
* **System Flow:**
    `Initial Path Resolution` -> `Depth-Aware Validation` -> `Secure Open (O_PATH)` -> `FD Pinning & Attestation` -> `FD Passing (Shadow-FS)` -> `Session-Bound I/O`
* **APIs / Interfaces:**
    * `FDPersistence.Pin(path)`: Resolves, validates, and pins an FD.
    * `FDPersistence.Get(handle)`: Returns a duplicated FD for a pinned handle.
* **Data Storage/State:**
    * Session-bound registry of `Handle -> {FD, Inode, AttestationToken}`.

## 5. Alternatives Considered
* **Continuous Re-Attestation**: Rejected due to high performance overhead and the remaining window for TOCTOU.
* **User-Space Shadow-FS Only**: Rejected as it still relies on the host OS's path resolution for the initial link, which can be raced.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The pinned FD is the "Root of Truth" for the session. Any mismatch in hardware Inode during subsequent checks triggers an immediate session termination.
* **Observability:** Logs all "Pinning Events" and "Blocked Race Attempts" with full FD metadata.

## 7. Evolutionary Changelog
* **2026-05-04:** Initial Document Creation.
* **2026-05-05: HEPA Evolution & Shared Memory Optimization**
    * **Context:** Market sync revealed the need for zero-latency path validation via hardware enclaves (HEPA).
    * **Architecture Adjustment:** Integrating `HEPA.ValidatePath()` into the `Initial Path Resolution` phase in Section 4.
    * **Optimization:** Introducing "FD-over-Shared-Memory" for the Zero-Copy transport, allowing pinned FDs to be shared with sub-millisecond latency between isolated RAMS shards.
