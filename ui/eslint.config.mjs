/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import nextPlugin from "@next/eslint-plugin-next";
import reactPlugin from "eslint-plugin-react";
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
      "react": reactPlugin,
      "react-hooks": reactHooksPlugin
    },
    rules: {
       ...tsPlugin.configs.recommended.rules,
       ...nextPlugin.configs.recommended.rules,
       ...reactHooksPlugin.configs.recommended.rules,
       "no-undef": "off",
       "@typescript-eslint/no-unused-vars": "off",
       "@typescript-eslint/no-explicit-any": "off",
       "@next/next/no-img-element": "off",
       "react-hooks/rules-of-hooks": "off",
       "react-hooks/exhaustive-deps": "off",
       "react/prop-types": "off",
       "react/no-unknown-property": "off",
       "react/no-unescaped-entities": "off",
       "react-hooks/set-state-in-effect": "off",
       "react-hooks/purity": "off",
       "react-hooks/set-state-in-render": "off",
       "react-hooks/immutability": "off"
    }
  },
  {
    ignores: [".next/**", "node_modules/**", "eslint.config.mjs", "next-env.d.ts", "dist/**", "test-results/**", "playwright-report/**"]
  }
];
