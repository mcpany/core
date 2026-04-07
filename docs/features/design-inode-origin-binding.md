# Design Doc: Deterministic Inode-to-Origin Binding
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
Recent post-mortems of the "ClawHavoc" crisis and subsequent security research into high-frequency agent coordination have identified a critical vulnerability: **Discovery-Phase Inode Racing**. Malicious subagents can perform Time-of-Check to Time-of-Use (TOCTOU) attacks on the filesystem, swapping a validated tool configuration for a malicious one in the milliseconds between validation and ingestion.

Deterministic Inode-to-Origin Binding solves this by cryptographically pinning the underlying hardware Inode to the initiating agent's origin at the moment of validation.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory `O_PATH` file descriptor pinning for all configuration files during the discovery phase.
    * Bind the pinned Inode to the hardware-attested session token of the initiating agent.
    * Neutralize TOCTOU filesystem attacks during tool discovery.
* **Non-Goals:**
    * Providing general-purpose filesystem encryption.
    * Replacing kernel-level sandbox drivers (it is a middleware implementation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Ensure that a newly discovered tool configuration is exactly the one that was validated.
* **The Happy Path (Tasks):**
    1. Agent initiates a tool discovery command.
    2. MCP Any intercepts the file access and performs initial security validation.
    3. Simultaneously, MCP Any opens the file with `O_PATH` and retrieves its hardware Inode.
    4. The Inode is cryptographically hashed with the Agent's session token and the initiating Origin.
    5. During the ingestion phase, MCP Any verifies that the file handle still points to the same pinned Inode.
    6. If a mismatch is detected (indicating an Inode race), the discovery is aborted and an alert is issued.

## 4. Design & Architecture
* **System Flow:**
    `Discovery Request` -> `Validation` -> `O_PATH Pinning` -> `Inode Binding` -> `Ingestion Verification`
* **APIs / Interfaces:**
    * `InodeBinder`: `PinFile(path string, sessionToken string) (PinnedInode, error)`
    * `InodeBinder`: `VerifyBinding(pinnedInode PinnedInode) error`
* **Data Storage/State:**
    * Transient map of Pinned Inodes to Session Tokens stored in kernel-bound memory.

## 5. Alternatives Considered
* **Full Path Monitoring (inotify)**: Rejected due to race conditions in high-frequency environments where the attacker can act faster than the notification.
* **Static File Hashing**: Rejected because it doesn't prevent file swapping if the swap happens *after* the hash is calculated but before the agent reads the content.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Part of the "Deterministic Attestation Gateway" initiative.
* **Observability:** Integrated with the "Local Security Violation Monitor" to flag Inode racing attempts.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
