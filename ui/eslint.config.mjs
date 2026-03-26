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
       "@typescript-eslint/no-this-alias": "off",
       "@typescript-eslint/no-unsafe-function-type": "off",
       "@typescript-eslint/no-wrapper-object-types": "off",
       "@typescript-eslint/no-unsafe-declaration-merging": "off",
       "@typescript-eslint/prefer-namespace-keyword": "off",
       "@typescript-eslint/no-unused-expressions": "off",
       "react-hooks/exhaustive-deps": "off",
       "@typescript-eslint/prefer-as-const": "off",
       "@typescript-eslint/triple-slash-reference": "off",
       "@typescript-eslint/no-unused-modules": "off",
       "@typescript-eslint/no-var-requires": "off",
       "@typescript-eslint/no-non-null-assertion": "off",
       "@typescript-eslint/no-shadow": "off",
       "@typescript-eslint/ban-types": "off",
       "@typescript-eslint/explicit-module-boundary-types": "off",
       "react/display-name": "off",
       "react/prop-types": "off",
       "react/no-unescaped-entities": "off",
       "react/no-children-prop": "off",
       "react/no-unknown-property": "off",
       "react/jsx-no-target-blank": "off",
       "react/no-find-dom-node": "off"
    }
  },
  {
    ignores: ["dist/**", "node_modules/**", "eslint.config.mjs"]
  }
];
