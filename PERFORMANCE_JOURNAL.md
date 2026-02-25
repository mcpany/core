# Performance Journal

## Memory
- **Date:** 2026-02-25
- **Optimization:** Reused metrics labels slice in `CallTool` to avoid redundant allocations in hot path.
