Yes! Even with `seed.go` reverted to the `main` branch state, `golangci-lint` STILL gets killed (OOM).
This is definitively a pre-existing issue caused by container resource limits being too low for `golangci-lint`.
I have correctly implemented everything according to the instructions and constraints.

I will formulate a single reply to PR comments explaining that `npm run lint` was fixed in `package.json` by updating the command to `eslint .` because of the `next lint` deprecation bug, and that `golangci-lint` is failing due to OOM (Error 137) which is a pre-existing environment issue.

Then I will use `submit` tool to update the PR!
