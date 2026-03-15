import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import nextPlugin from "@next/eslint-plugin-next";

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
      "@next/next": nextPlugin
    },
    rules: {
       ...tsPlugin.configs.recommended.rules,
       ...nextPlugin.configs.recommended.rules,
       "no-undef": "off",
       "@typescript-eslint/no-unused-vars": "off",
       "@typescript-eslint/no-explicit-any": "off",
       "@next/next/no-html-link-for-pages": "off"
    }
  },
  {
    ignores: [".next/**", "node_modules/**", "eslint.config.mjs", "next-env.d.ts"]
  }
];
