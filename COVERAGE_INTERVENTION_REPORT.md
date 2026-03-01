# Coverage Intervention Report

* **Target:** `server/pkg/util/json_size.go`
* **Risk Profile:** This module handles custom JSON size estimation to avoid large string allocations, implementing a highly recursive structure with many fast paths. It touches numerous data types and reflection boundaries (`reflect.Value`), putting it at a high risk for edge-case errors like panics or infinite recursion in the absence of sufficient test coverage. Prior to this intervention, its statement coverage was ~53%, with crucial paths related to array sizing and boolean checking entirely untested.
* **New Coverage:**
  * Slices, empty interfaces, zero-valued variables.
  * Explicit tests for different `uint` boundaries (`uint8`, `uint16`, `uint32`, `uint64`).
  * Explicit float definitions and integer sizing boundary testing.
  * Empty map and slice testing to guard against panics or map allocations.
  * Structs featuring the `omitempty` tag.
  * Cyclic struct pointers and slice pointers to prevent stack overflows, proving the cyclic safety features using a memory-pool based recursive tracker.
* **Verification:** `make lint` passes. Unit tests pass cleanly. Statement coverage for `json_size.go` specifically increased from ~53% to ~88%, guarding the most critical and complex logic paths.
