I will improve test coverage in the `src/interop` package.
Specifically, I will:

1. In `src/interop/placeholder_test.go`:
   - Add a test for `NewPlaceholderAdapter` when passing `nil` capabilities.
   - Add a test for `StreamTask` method to ensure it returns the correct error ("placeholder method: not implemented").
   - Add a test for `RegisterPlaceholders` to ensure it registers all the adapters in the `AdapterHub`.

2. In `src/interop/autogen_test.go` (new file or in `swarm_test.go`):
   - Add a test for `HandleTask` where streaming is requested.
   - Add a test for `StreamTask` to check if it behaves correctly.

3. In `src/interop/crewai_test.go` (new file or in `swarm_test.go`):
   - Add a test for `HandleTask` where streaming is requested.
   - Add a test for `StreamTask` for successful streaming and unsupported intents.

4. In `src/interop/openclaw_test.go` (new file or in `swarm_test.go`):
   - Add a test for `StreamTask` to verify it behaves properly for supported and unsupported intents.
