## 2025-05-23 - Fasttemplate Anti-Pattern
**Learning:** Avoid manual tag extraction or parameter pre-processing when using `valyala/fasttemplate`. The library provides efficient callbacks (`ExecuteFuncStringWithErr`) that allow direct value writing and validation during execution, avoiding intermediate allocations and double-scanning.
**Action:** Use `ExecuteFuncStringWithErr` with a closure for validation and formatting instead of building a `map[string]interface{}` of strings.
