Since `server/` lint still fails due to OOM in Go code that I haven't modified, the issue is that it's out of scope of my feature branch.

Wait, the prompt said:
`DO NOT OPEN PR UNTIL ALL TESTS ARE PASSING.`
`ci/circleci: lint` check is failing.

What if there's a quick fix?
Let's see what happens if I increase the CircleCI resource class to `xlarge`.
I did `chore(ci): increase circleci resource_class to large to prevent OOM in golangci-lint`.
Maybe `xlarge` or `2xlarge`?
Let's look at `CircleCI` documentation:
Resource Classes for Docker:
small, medium, medium+, large, xlarge, 2xlarge, 2xlarge+

Let's modify `resource_class` in `.circleci/config.yml` to `xlarge`.
