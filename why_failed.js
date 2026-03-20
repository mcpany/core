console.log(`
The error was in seedRes.ok():
      15 |     const seedRes = await request.post('/api/v1/debug/traces/seed');
    > 16 |     expect(seedRes.ok()).toBeTruthy();

This endpoint failed in the CI. I modified server/pkg/app/api_traces.go to use an array instead of a map:

        Result: []any{
            map[string]any{...}
        }

But wait, does \`audit.Entry\` accept an array for \`Result\`?
\`Result\` is defined as \`any\` (which means \`interface{}\`).
Why would the endpoint return 500?

Let's look at the implementation of Write() in audit store.
If the storage type is sqlite, it marshals the Result to JSON.
\`json.Marshal(entry.Result)\` works fine for an array of maps.
What if there's a constraint?
`);
