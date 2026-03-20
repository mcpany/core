import subprocess
import time

print("Starting backend...")
backend = subprocess.Popen(["go", "run", "cmd/server/main.go", "run", "--config-path", "config.minimal.yaml"], cwd="server", env={"GOPATH": "/app/build/env", "GOMODCACHE": "/app/build/env/pkg/mod", "MCPANY_API_KEY":"test-token", "PATH": "/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"})

print("Waiting for backend...")
time.sleep(15)

print("Starting frontend...")
frontend = subprocess.Popen(["npm", "run", "dev"], cwd="ui")

print("Waiting for frontend...")
time.sleep(10)

print("Running test...")
subprocess.run(["python", "/home/jules/verification/verify_inspector.py"])

backend.terminate()
frontend.terminate()
