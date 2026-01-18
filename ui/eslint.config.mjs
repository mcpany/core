/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import nextPlugin from "@next/eslint-plugin-next";
import reactHooksPlugin from "eslint-plugin-react-hooks";

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
      "@next/next": nextPlugin,
      "react-hooks": reactHooksPlugin
    },
    rules: {
       ...tsPlugin.configs.recommended.rules,
       ...nextPlugin.configs.recommended.rules,
       "react-hooks/rules-of-hooks": "error",
       "react-hooks/exhaustive-deps": "warn",
       "no-undef": "off",
       "@typescript-eslint/no-unused-vars": ["warn", { "argsIgnorePattern": "^_", "varsIgnorePattern": "^_", "caughtErrorsIgnorePattern": "^_" }],
       "@typescript-eslint/no-explicit-any": "warn"
    }
  },
  {
    ignores: [".next/**", "node_modules/**", "eslint.config.mjs", "next-env.d.ts"]
  }
];
