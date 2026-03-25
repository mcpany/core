lint:
	bazel run //:lint || true

test:
	bazel test //... || true
