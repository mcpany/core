const { execSync } = require('child_process');

try {
  console.log("Starting backend server...");
  const backend = execSync("make run", { cwd: "/app/server", timeout: 10000 });
} catch (e) {
  console.log("Backend failed or timed out", e.message);
}
