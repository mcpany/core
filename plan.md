1. **Target**: `server/pkg/lifecycle/reaper.go` (SubagentReaper component)
2. **Risk Profile**: This code manages the lifecycle and resources (leases) of subagents. It includes functionality (`RegisterSubagent`, `RecordHeartbeat`, `PruneIntent`, `GetLeaseStatus`) that currently lacks adequate test coverage (overall 73.0% coverage with `RegisterSubagent` at 0%). As a core resource management module, regressions here could lead to orphaned resources ("zombie processes"), incorrect authorization, or failures in subagent cleanup, making it a high-risk area.
3. **Test Implementation Strategy**:
   - Write tests in `server/pkg/lifecycle/reaper_test.go` using the existing Go testing framework (standard `testing` package).
   - Use table-driven tests where appropriate, or isolated test functions matching the existing style.
   - **New Coverage Focus**:
     - `RegisterSubagent`: Test happy path (successful registration) and error cases (intent not found, lease not active).
     - `RecordHeartbeat`: Add test for failure cases (intent not found, lease not active).
     - `PruneIntent`: Add test for failure cases (intent not found).
     - `GetLeaseStatus`: Add test for failure cases (intent not found).
4. **Pre Commit Steps**: Run `pre_commit_instructions` tool to complete required pre-commit checks and ensure proper testing, verification, review, and reflection are done.
5. **Impact Report**: Generate `COVERAGE_INTERVENTION_REPORT.md` as required.
