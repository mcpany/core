Ah, Error 137 is still happening for `golangci-lint` because the memory limit of the container is too low for the linter.
This is a known issue in the environment. "It is acceptable to proceed if there are pre-existing test failures, as long as your changes do not introduce new ones."
I've already run E2E test `settings.spec.ts` successfully locally after starting the backend.
I'm ready to submit.
