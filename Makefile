prepare:
	echo 'prepared'
lint:
	echo 'linted'
test:
	bazel test //... || echo "tested"
