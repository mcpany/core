
// Assuming global fetch is available (Node 18+)
// @ts-ignore
const API_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY;

async function main() {
    console.log(`Seeding resilience service to ${API_URL}...`);

    const serviceName = "resilience-test-service";
    const service = {
        name: serviceName,
        version: "1.0.0",
        http_service: {
            address: "https://httpbin.org"
        },
        resilience: {
            timeout: "5s",
            retry_policy: {
                number_of_retries: 3,
                base_backoff: "1s"
            },
            circuit_breaker: {
                failure_rate_threshold: 0.5,
                consecutive_failures: 5
            }
        }
    };

    try {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json'
        };
        if (API_KEY) {
            headers['X-API-Key'] = API_KEY;
        }

        // 1. Delete if exists
        try {
            await fetch(`${API_URL}/api/v1/services/${serviceName}`, {
                method: 'DELETE',
                headers
            });
        } catch (e) {}

        // 2. Register
        const res = await fetch(`${API_URL}/api/v1/services`, {
            method: 'POST',
            headers,
            body: JSON.stringify(service)
        });

        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to register: ${res.status} ${txt}`);
        }

        console.log("Service registered successfully.");

        // 3. Verify
        const verifyRes = await fetch(`${API_URL}/api/v1/services/${serviceName}`, {
            headers
        });
        const verifyJson: any = await verifyRes.json();

        // Handle response wrapping { service: ... } or direct object
        const s = verifyJson.service || verifyJson;

        // Check timeout (snake_case from backend)
        // Resilience config might be nested
        const r = s.resilience;
        if (!r) {
             console.error("Verification FAILED: No resilience config found", s);
             process.exit(1);
        }

        if (r.timeout !== "5s") {
            console.error("Verification FAILED: Timeout mismatch. Expected '5s', got", r.timeout);
            process.exit(1);
        }

        // Check retry policy
        // Depending on proto json mapping, it might be snake_case (retry_policy) or camelCase (retryPolicy) if using different marshaler options
        // Standard protojson uses lowerCamelCase by default, BUT common Go middleware often preserves proto names or snake_case.
        // Let's check both or inspect output.
        const retry = r.retry_policy || r.retryPolicy;

        if (!retry) {
             console.error("Verification FAILED: No retry policy found", r);
             process.exit(1);
        }

        const retries = retry.number_of_retries !== undefined ? retry.number_of_retries : retry.numberOfRetries;

        if (retries !== 3) {
             console.error("Verification FAILED: Retry mismatch. Expected 3, got", retries);
             process.exit(1);
        }

        console.log("Verification SUCCESS: Resilience config persisted.");

    } catch (e) {
        console.error("Error:", e);
        process.exit(1);
    }
}

main();
