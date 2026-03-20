const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

// In my earlier replacement I replaced:
// content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "");
// However, looking at git diff there was no difference?
// Ah, git diff won't show it if the file was modified without `git add` but `git diff` should show it. Wait, `ui/tests` might be ignored?
// No, it's not ignored.
// Let's actually look at the server logs to see why `/api/v1/debug/seed` returns 500.
// Is the server running during tests? Yes, it's handled by bazel.
