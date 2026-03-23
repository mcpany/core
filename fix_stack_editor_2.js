const fs = require('fs');

let code = fs.readFileSync('ui/tests/stack-editor.spec.ts', 'utf-8');

code = code.replace(
    /const weatherNode = visualizer\.locator\('\.react-flow__node'\)\.filter\(\{ hasText: 'weather-service' \}\)\.first\(\);\n\s*await expect\(weatherNode\)\.toBeVisible\(\{ timeout: 15000 \}\);\n\s*await expect\(weatherNode\)\.toContainText\('weather-service'\);/,
    `// For a new stack, there's no pre-populated node.
      // But we can insert text into Monaco to see it render
      const editorTextarea = page.locator('.monaco-editor textarea');
      await editorTextarea.focus();
      await page.keyboard.press('Meta+a');
      await page.keyboard.press('Control+a');
      await page.keyboard.press('Backspace');
      const newYaml = \`name: \${stackName}\\nversion: "1.0"\\nservices:\\n  - name: weather-service\\n    command_line_service:\\n      command: "echo weather"\\n\`;
      await page.keyboard.insertText(newYaml);

      const weatherNode = visualizer.locator('.react-flow__node').filter({ hasText: 'weather-service' }).first();
      await expect(weatherNode).toBeVisible({ timeout: 15000 });
      await expect(weatherNode).toContainText('weather-service');`
);

fs.writeFileSync('ui/tests/stack-editor.spec.ts', code);
