import subprocess
import time
import os
import signal
import sys
from playwright.sync_api import sync_playwright

def verify_stacks():
    # Setup paths
    repo_root = os.getcwd()
    server_bin = os.path.join(repo_root, "build/bin/server")
    server_config = os.path.join(repo_root, "server/config.minimal.yaml")
    ui_dir = os.path.join(repo_root, "ui")

    print(f"Starting backend server from {server_bin} with config {server_config}...")
    backend_process = subprocess.Popen(
        [server_bin, "run", f"--config-path={server_config}"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=repo_root
    )

    # Wait for backend to be reasonably up (simple sleep for now)
    time.sleep(5)
    if backend_process.poll() is not None:
        print("Backend failed to start!")
        print(backend_process.stdout.read().decode())
        print(backend_process.stderr.read().decode())
        sys.exit(1)

    print("Starting frontend server...")
    # Using 'npm start' requires 'npm run build' which we did earlier
    frontend_process = subprocess.Popen(
        "npm start",
        shell=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=ui_dir,
        env={**os.environ, "PORT": "3000"}
    )

    # Wait for frontend to be up
    time.sleep(15)
    if frontend_process.poll() is not None:
        print("Frontend failed to start!")
        print(frontend_process.stdout.read().decode())
        print(frontend_process.stderr.read().decode())
        backend_process.terminate()
        sys.exit(1)

    print("Running Playwright verification...")
    screenshot_path = "/home/jules/verification/stacks_verification.png"
    os.makedirs(os.path.dirname(screenshot_path), exist_ok=True)

    browser = None
    try:
        with sync_playwright() as p:
            browser = p.chromium.launch(headless=True)
            page = browser.new_page()

            # Go to stacks page
            url = "http://localhost:3000/stacks"
            print(f"Navigating to {url}")
            try:
                page.goto(url, timeout=30000)
            except Exception as e:
                print(f"Failed to load page: {e}")
                # Print output to debug
                # Kill processes
                frontend_process.terminate()
                backend_process.terminate()
                # Read output
                print("Frontend Output:")
                print(frontend_process.stdout.read().decode())
                print(frontend_process.stderr.read().decode())
                print("Backend Output:")
                print(backend_process.stdout.read().decode())
                print(backend_process.stderr.read().decode())
                sys.exit(1)

            # Wait for content to load
            try:
                page.wait_for_load_state("networkidle", timeout=10000)
            except Exception as e:
                print(f"Warning: networkidle timeout: {e}")

            # Take screenshot
            print(f"Taking screenshot to {screenshot_path}")
            page.screenshot(path=screenshot_path)

            # Check for specific text to ensure page loaded correctly
            content = page.content()
            if "Stacks" in content:
                print("SUCCESS: Found 'Stacks' text on page.")
            else:
                print("WARNING: 'Stacks' text not found on page. Check screenshot.")
                print(f"Page Content: {content[:500]}...")

            browser.close()

    except Exception as e:
        print(f"Error during verification: {e}")
        if browser:
            browser.close()
    finally:
        print("Cleaning up processes...")
        backend_process.terminate()
        frontend_process.terminate()
        backend_process.wait()
        frontend_process.wait()

if __name__ == "__main__":
    verify_stacks()
