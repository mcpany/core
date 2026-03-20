const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

// If there's an error from the backend, let's see it in the error message.
// The code already does `throw new Error(\`Failed to seed global state: \${res.status()} \${text}\`);` but text was probably empty string.
// Let's modify our spec to log it.
