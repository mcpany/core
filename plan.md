Wait, CI Checks Failed!
The lint check failed:
`ci/circleci: lint`

Let's check the build output and fix the lint issues.
Earlier `npm run lint` gave this error:
`next: not found` because it wasn't run via `npm run lint`. Wait, I did `cd ui && npm run lint`.
It said:
```
 `next lint` is deprecated and will be removed in Next.js 16.
For new projects, use create-next-app to choose your preferred linter.
For existing projects, migrate to the ESLint CLI:
npx @next/codemod@canary next-lint-to-eslint-cli .

Invalid Options:
- Unknown options: useEslintrc, extensions, resolvePluginsRelativeTo, rulePaths, ignorePath, reportUnusedDisableDirectives
```
This means `npm run lint` in Next.js is failing due to some invalid ESLint options in `eslint.config.mjs` or `.eslintrc.json`.
But wait! There was ALSO a Go lint error:
`make build` failed with `golangci-lint` errors? No, `make build` succeeded recently after I fixed the protobuf imports.
Let's run `make lint` in the root directory to see what failed.
