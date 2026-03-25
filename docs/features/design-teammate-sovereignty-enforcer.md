# Design Doc: Teammate Sovereignty Enforcer (TSE)
**Status:** Draft | **Created:** 2026-06-18

## 1. Context and Scope
Parallel teammates in agent teams require cryptographically isolated "Sovereignty Shards" to prevent "State Smearing" and logic propagation.

## 2. Goals
- Provide cryptographically bound isolation.
- Enforce "Auth-at-the-Pipe" for inter-agent comms.
- Bind identities to hardware-attested tokens.

## 3. Critical User Journey
1. Lead Agent initializes mission with TPM token.
2. Lead spawns teammates; TSE issues isolated shards.
3. Comms occur via isolated named pipes managed by the gateway.

## 4. Architecture
- Isolation Kernel: Manages shard lifecycles.
- Logic-Aware Shard Guards: Filters reasoning fragments via LSV.

## 7. Evolutionary Changelog
- **2026-06-18:** Initial Document Creation.
### Update: 2026-06-18 - Logic-Path Interdiction Integration
**Architecture Adjustment:** Mandatory integration with LSV.
