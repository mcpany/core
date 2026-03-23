# Design Doc: Attention-Locked Sovereignty (ALS) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Mitigates CVE-2026-71002 (Context-Window Ghosting). Locks critical fragments from subagent shadowing.

## 2. Goals & Non-Goals
Goals: Fragment sealing, Boundary enforcement.

## 3. CUJ
Parent locks context -> Interception -> Sanitized view for subagent.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
