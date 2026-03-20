const fs = require('fs');

// We previously found the CI test failed because we used an array of maps for Result.
// Let's modify the API trace seed to just use a regular struct for Result as previously defined,
// or a simple map that contains `data` array, because the `Result` in `audit.Entry` is `any`.
// If `Result` is `any`, sqlite stores it by `json.Marshal(entry.Result)`.
// We had:
// 			Result: map[string]any{
// 				"content": []any{ ... }
//			}
// Wait, the failure was actually on the FIRST run:
//       14 |     // Call the /api/v1/debug/traces/seed endpoint to populate the DB with a rich trace
//       15 |     const seedRes = await request.post('/api/v1/debug/traces/seed');
//     > 16 |     expect(seedRes.ok()).toBeTruthy();
// This returned 500.

// Let's fix the trace seed to match what works.
// Before my changes, it was:
// 			Result: map[string]any{
// 				"summary":    "Revenue up 15%",
// 				"confidence": 0.98,
// 			},
//
// If I use an array of objects inside a map, it should be fine. But what if it wasn't valid Go?
