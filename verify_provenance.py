from playwright.sync_api import sync_playwright, expect

def verify_service_provenance():
    with sync_playwright() as p:
        # Launch browser
        browser = p.chromium.launch(headless=True)
        # Create context with ignore_https_errors to handle self-signed certs in dev
        context = browser.new_context(ignore_https_errors=True)
        page = context.new_page()

        try:
            # Navigate to the service details page
            # We assume the service 'verified-service' has been seeded by the test runner or manually
            # But the backend might not be running with the seeded data in this verification context.
            # We need to rely on what `make run` provides or seed it via API first.

            # Since I can't easily restart the backend with seeded data here in python script without complex setup,
            # I will rely on the `ui/tests/supply_chain.spec.ts` which I already verified.
            # However, for visual verification, I need the UI running.
            # Assuming 'make run' starts the UI at localhost:3000 and Server at localhost:8080.

            # Seed data via API first (Server must be running)
            api_url = "http://localhost:50050/api/v1/services" # Default server port

            # If server is not running, this script will fail, which is expected behavior for verification instructions.
            # I will assume the server is running or I should start it.
            # The instructions say "Start the Application". I will do that in bash separately.

            # 1. Seed Verified Service
            page.request.post(api_url, data={
                "name": "verified-service-ui-test",
                "id": "verified-service-ui-test",
                "version": "1.0.0",
                "http_service": {"address": "https://example.com"},
                "provenance": {
                    "verified": True,
                    "signer_identity": "Google Cloud",
                    "attestation_time": "2024-01-01T12:00:00Z",
                    "signature_algorithm": "ECDSA-P256"
                }
            })

            # 2. Navigate to UI
            page.goto("http://localhost:3000/upstream-services/verified-service-ui-test")

            # 3. Click Supply Chain tab
            page.get_by_role("tab", name="Supply Chain").click()

            # 4. Wait for content
            expect(page.get_by_text("Verified Service", exact=True)).to_be_visible()
            expect(page.get_by_text("Google Cloud")).to_be_visible()

            # 5. Take Screenshot
            page.screenshot(path="/tmp/verification_provenance.png")
            print("Screenshot taken at /tmp/verification_provenance.png")

        except Exception as e:
            print(f"Verification failed: {e}")
            # Take screenshot anyway for debug
            page.screenshot(path="/tmp/verification_failure.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_service_provenance()
