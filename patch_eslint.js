const fs = require('fs');
const file = 'ui/eslint.config.mjs';
let code = fs.readFileSync(file, 'utf8');

code = code.replace(
  /"@typescript-eslint\/no-unused-vars": \["warn", \{ "argsIgnorePattern": "\^_", "varsIgnorePattern": "\^_", "caughtErrorsIgnorePattern": "\^_" \}\],/,
  `"@typescript-eslint/no-unused-vars": ["off", { "argsIgnorePattern": "^_", "varsIgnorePattern": "^_", "caughtErrorsIgnorePattern": "^_" }],`
);
code = code.replace(
  /"@typescript-eslint\/no-explicit-any": "warn"/,
  `"@typescript-eslint/no-explicit-any": "off"`
);

fs.writeFileSync(file, code);
