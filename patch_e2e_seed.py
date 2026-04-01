import sys

with open("ui/tests/e2e/test-data.ts", "r") as f:
    content = f.read()

search = """        rootSpan: {
            id: 'span-1',
            name: 'calculate_sum',
            serviceName: 'Math',
            type: 'tool',
            status: 'success',
            startTime: Date.now() - 150,
            endTime: Date.now(),
            children: []
        },"""

replace = """        rootSpan: {
            id: 'span-1',
            name: 'calculate_sum',
            serviceName: 'Math',
            type: 'tool',
            status: 'success',
            startTime: Date.now() - 150,
            endTime: Date.now(),
            input: {
                query: "Analyze Q3 financial report",
                context: {
                    session_id: "user-session-123",
                    flags: ["fast", "experimental"],
                    settings: {
                        timeout_ms: 5000,
                        retry: true,
                        max_retries: 3,
                        null_val: null
                    }
                }
            },
            output: {
                summary: "Revenue up 15%",
                confidence: 0.98,
                metadata: {
                    processed_at: "2023-10-27T10:00:00Z",
                    sources: [
                        { id: "src-1", type: "pdf", pages: 15 },
                        { id: "src-2", type: "database", rows_scanned: 10500 }
                    ],
                    tags: ["finance", "q3", "internal"]
                }
            },
            children: []
        },"""

content = content.replace(search, replace)

with open("ui/tests/e2e/test-data.ts", "w") as f:
    f.write(content)
