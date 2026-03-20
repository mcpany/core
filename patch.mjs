import fs from 'fs';

const file = 'ui/eslint.config.mjs';
let data = fs.readFileSync(file, 'utf8');

data = data.replace('];', `  },
  {
    files: ["**/*.ts", "**/*.tsx"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off"
    }
  }
];`);

fs.writeFileSync(file, data);
