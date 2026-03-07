# Design Doc: CAS-Based Skill Installer
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
The recent critical vulnerabilities in OpenClaw (CVE-2026-28486, CVE-2026-28456) have exposed the dangers of traditional "install-to-disk" patterns for AI agent tools. Path traversal attacks (Zip Slip) and uncontrolled search paths allow malicious tools to escape their intended directories and compromise the host system. MCP Any must provide a secure, immutable way to install and run MCP servers.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a Content-Addressed Storage (CAS) system for all local MCP server installations.
    * Ensure strict filesystem isolation for installed tools.
    * Provide cryptographic verification of all installed artifacts.
    * Enable atomic updates and easy rollbacks.
* **Non-Goals:**
    * Creating a full containerization engine (like Docker).
    * Managing remote MCP server deployment (only local installs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Install a new MCP server from a community repository without risking host filesystem compromise.
* **The Happy Path (Tasks):**
    1. User provides a URL or archive for a new MCP server.
    2. MCP Any downloads and extracts the archive into a memory-backed staging area.
    3. The installer validates all file paths (rejecting any traversals) and calculates SHA-256 hashes for every file.
    4. Files are moved into a CAS structure (e.g., `storage/blobs/sha256/...`).
    5. A manifest is created linking the tool to its specific CAS-pinned files.
    6. When executed, the tool is provided a restricted "View" of its files via a Virtual Filesystem (VFS) or bind-mount.

## 4. Design & Architecture
* **System Flow:**
    `Archive -> Staging -> Path/Hash Validation -> CAS (Storage) -> Tool Manifest -> VFS Execution`
* **APIs / Interfaces:**
    * `mcpany install <url/path>`: CLI command for secure installation.
    * `SkillManager` service: Internal Go service managing the CAS and manifests.
* **Data Storage/State:**
    * `storage/blobs/`: Content-addressed file storage.
    * `storage/manifests/`: JSON/Protobuf files describing tool structures and metadata.

## 5. Alternatives Considered
* **Standard Directory Isolation**: Simply installing to `/opt/mcpany/tools/tool-name`. *Rejected* as it is still vulnerable to misconfigured permissions and symlink attacks.
* **Docker-only Installation**: Requiring all tools to be Docker images. *Rejected* due to high overhead and complexity for simple script-based tools.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CAS ensures immutability. Path validation prevents Zip Slip. VFS prevents unauthorized read/write to the host.
* **Observability:** Audit logs will record every file hash installed and every execution attempt.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
