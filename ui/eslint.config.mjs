/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";

export default [
  {
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaFeatures: { modules: true, jsx: true },
        sourceType: "module",
        ecmaVersion: "latest"
      }
    },
    plugins: {
      "@typescript-eslint": tsPlugin,
    },
    rules: {
       "no-undef": "off",
       "@typescript-eslint/no-unused-vars": "off",
       "@typescript-eslint/no-explicit-any": "off",
       "@typescript-eslint/ban-ts-comment": "off",
       "@typescript-eslint/no-require-imports": "off",
       "@typescript-eslint/no-empty-object-type": "off",
       "@typescript-eslint/no-unused-expressions": "off",
       "@typescript-eslint/no-non-null-asserted-optional-chain": "off",
       "@typescript-eslint/no-this-alias": "off"
    }
  },
  {
    ignores: ["dist/**", "node_modules/**", "eslint.config.mjs"]
  }
];
