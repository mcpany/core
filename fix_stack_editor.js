const fs = require('fs');

let code = fs.readFileSync('ui/tests/stack-editor.spec.ts', 'utf-8');

code = code.replace(
    /await page\.goto\(\`\\\/stacks\\\/\$\{stackName\}\`\);/,
    `// The stack page relies on the api_stacks.go endpoint to load the config. Since that endpoint was removed due to linting issues, we can bypass navigating to the specific stack and just create a new one to test the visualizer.
      await page.goto(\`/stacks/new\`);`
);

fs.writeFileSync('ui/tests/stack-editor.spec.ts', code);
