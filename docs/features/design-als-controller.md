# Design Doc: ALS Controller
**Status:** Draft | 2026-06-18
## 1. Context and Scope
Prevents Context-Window Ghosting.
## 2. Goals & Non-Goals
* Goals: Enforce token quotas.
## 3. Critical User Journey
1. Subagent reaches quota. 2. ALS terminates loop.
## 4. Design & Architecture
Middleware based token tracking.
## 5. Alternatives Considered
Model-native gating (too opaque).
## 6. Concerns
Performance overhead of token counting.
## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
