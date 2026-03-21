Wait, the `go.mod` is getting modified by something.
Let's undo `go mod edit -replace`.
```bash
git checkout server/go.mod server/go.sum
```
Then I will submit. I fixed the code in the UI and the backend seeder.
