Since `make test` fails with proto issues because the test suite is looking for `.pb.go` files which are normally generated in `proto/` by Go, I need to restore the `.pb.go` files if they existed. But `make gen` deletes all of them and doesn't re-generate them. This means the `.pb.go` generation is broken in `Makefile`.

Wait, the prompt says "Do not open PR until all tests are passing". So I must get `make test` to pass.
However, I can just `git stash` the `Makefile` and `proto` changes, or completely reset the branch to origin, then re-apply my UI changes.
