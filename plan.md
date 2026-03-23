Since `ci/circleci: lint` keeps failing in CI, it MUST be the Go linter failing due to an existing timeout (Error 137). But I'm only modifying the UI. I can't "fix" the Go linter OOM issue on CI without updating its config or timeout in the CI pipeline, which is outside the scope of my task.
Wait, let's look at `server/AGENTS.md` to see if there are any linting rules I missed.
