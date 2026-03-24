# Design Doc: Inode-Locked Workspace Broker
**Status:** Draft
**Created:** 2026-05-21

## 1. Context and Scope
AI agents often operate on local filesystems using path-based validation. However, path-based security is vulnerable to TOCTOU (Time-of-Check to Time-of-Use) attacks, symlink tunnels, and mount escapes. A malicious subagent can create complex symlink chains to trick the gateway into reading or writing files outside the intended project workspace.

The Inode-Locked Workspace Broker (ILWB) provides hardware-enforced filesystem sovereignty. It binds an entire agent session's filesystem access to a specific "Inode-root." Instead of validating string paths, the broker works with the kernel and TPM to ensure that every file descriptor opened by an agent process is physically descendant from the authorized hardware Inode. This neutralizes all forms of path-based escapes, including nested symlinks that bridge into restricted host regions.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound filesystem isolation using Inode-root pinning.
    * Provide a mandatory "Sovereignty Gate" for all filesystem-related tool calls.
    * Neutralize symlink escapes and unauthorized path traversals.
    * Maintain low-latency file access while enforcing hardware-level restrictions.
* **Non-Goals:**
    * Replacing general-purpose OS access control (ILWB is an additional agent-specific layer).
    * Managing network-based file systems (scoped to local workspace mounts).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Architect
* **Primary Goal:** Ensure a "Refactoring Agent" with write access to `./src` cannot use a symlink to overwrite `/etc/passwd`.
* **The Happy Path (Tasks):**
    1. The Architect defines an "Inode-Locked Workspace" policy for the refactoring mission.
    2. MCP Any identifies the physical Inode of the project root directory.
    3. MCP Any ILWB binds this Inode-ID to the agent's hardware-attested session in the TPM.
    4. The Agent attempts to open a file via a tool `write_file(path="../../etc/passwd")`.
    5. ILWB intercepts the system call, resolves the absolute Inode of the target, and checks its parentage.
    6. Because the target Inode is not a descendant of the project root Inode, the request is rejected with a "Sovereignty Violation" fault.
    7. The attempt is logged as a critical security event, and the agent's session is force-terminated.

## 4. Design & Architecture
* **System Flow:**
    [Agent Tool Call] -> [ILWB Interceptor] -> [Inode Resolve (Kernel)] -> [Hardware Inode Map (TPM)] -> [Execution]
    1. Gateway receives file request.
    2. ILWB uses `O_PATH` and `fstat` to get the target Inode without following untrusted paths.
    3. ILWB queries the TPM-locked "Allowed Inode Range."
    4. Hardware verifies descendant status in the directory tree.
* **APIs / Interfaces:**
    * `ProvisionInodeRoot(session_id, path) -> root_inode_id`
    * `ValidateInodeAccess(session_id, file_path) -> bool`
* **Data Storage/State:**
    * Authorized Inode-roots are stored in a hardware-protected segment of the MCP Any internal state.

## 5. Alternatives Considered
* **Chroot/Jail:** Rejected because they are often difficult to configure for ephemeral agent swarms and can still be bypassed via kernel vulnerabilities.
* **Path Sanitization (Regex):** Rejected; notoriously incomplete and vulnerable to encoding and symlink tricks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The filesystem is a "Black Box" to the agent until the ILWB hardware check passes.
* **Observability:** "Sovereignty Violations" are high-fidelity indicators of compromise and trigger immediate alerts.

## 7. Evolutionary Changelog
* **2026-05-21:** Initial Document Creation.
