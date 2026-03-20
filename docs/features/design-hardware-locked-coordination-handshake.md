# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Hardware-Locked Coordination Handshake (HLCH)

**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

As agent swarms transition to horizontal meshes, they are becoming vulnerable to Identity-Decay Attacks (IDA).

## 2. Goals & Non-Goals

* **Goals:**
    * Implement mandatory HLCH-compliant handshake.
    * Bind coordination fragments to hardware-attested tokens.
* **Non-Goals:**
    * Replacing transport-level encryption.

## 3. Critical User Journey (CUJ)

* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent subagent Identity-Decay.
* **The Happy Path (Tasks):**
    1. Parent Agent establishes mesh.
    2. HLCH initiates handshake.
    3. Teammates provide hardware-bound tokens.
    4. IDA mimicry detected and blocked.

## 4. Design & Architecture

* **System Flow:**
    [Request] -> [HLCH Middleware] -> [Validator] -> [TPM]

## 5. Alternatives Considered

* **Software JWTs:** Rejected due to IDA vulnerability.

## 6. Evolutionary Changelog

* **2026-06-14:** Initial Document Creation.
