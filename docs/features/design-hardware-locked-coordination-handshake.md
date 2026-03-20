# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Hardware-Locked Coordination Handshake (HLCH)
**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

As agent swarms transition to long-running horizontal meshes, they are becoming vulnerable to Identity-Decay Attacks (IDA).

## 2. Goals & Non-Goals

* **Goals:**
    * Implement a mandatory HLCH-compliant handshake.
    * Bind every coordination fragment to a hardware-attested session token.

* **Non-Goals:**
    * Replacing transport-level encryption (TLS/mTLS).

## 3. Critical User Journey (CUJ)

* **User Persona:** Swarm Security Architect

* **Primary Goal:** Prevent a long-running subagent from performing Identity-Decay.

## 4. Design & Architecture

* **System Flow:**
    [Request] -> [HLCH Middleware] -> [Validator] -> [TPM/Enclave]

## 5. Alternatives Considered

* **Software JWTs:** Rejected because they are vulnerable to IDA.

## 6. Evolutionary Changelog

* **2026-06-14:** Initial Document Creation.
