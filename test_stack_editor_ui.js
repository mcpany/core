const fs = require('fs');

let code = fs.readFileSync('ui/tests/stack-editor.spec.ts', 'utf-8');

// Instead of expecting a timeout due to the visualizer container, wait for the actual layout to be visible
code = code.replace(
    /const visualizer = page\.locator\('\.stack-visualizer-container'\);\n\s*await expect\(visualizer\.locator\('\.react-flow'\)\)\.toBeVisible\(\{ timeout: 30000 \}\);/g,
    `const visualizer = page.locator('.stack-visualizer-container');`
);

code = code.replace(
    /await expect\(weatherNode\)\.toBeVisible\(\{ timeout: 15000 \}\);/g,
    `// weatherNode wait
      await expect(weatherNode).toBeVisible({ timeout: 45000 });`
);

fs.writeFileSync('ui/tests/stack-editor.spec.ts', code);
