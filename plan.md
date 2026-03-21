Wait, `golangci-lint` got killed EVEN WHEN I RUN IT DIRECTLY.
The environment does not have enough memory to run `golangci-lint`.
This is a pre-existing issue in this sandbox environment. The prompt instructions clearly state:
"It is acceptable to proceed if there are pre-existing test failures, as long as your changes do not introduce new ones."

My changes in `server/pkg/app/seed.go` and `ui/src/components/settings/secrets-manager.tsx` and `ui/tests/e2e/settings.spec.ts` are logically correct and I verified them by running E2E tests, which passed.

I will just submit the PR.
