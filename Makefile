prepare:
	echo 'prepared'
lint:
	bazelisk run //:lint

docker-lint:
	bazelisk run //:lint

docker-test:
	# Run tests in batches to avoid OOM
	bazelisk test //server/... --test_output=errors --config=local --jobs=4
	UI_TESTS=$$(bazelisk query "kind(test, //ui:all)" | grep "vitest"); \
	if [ -n "$$UI_TESTS" ]; then \
		echo "$$UI_TESTS" | xargs -n 10 bazelisk test --test_output=errors --config=local --jobs=4; \
	fi

k8s-e2e:
	# Run Playwright E2E tests
	E2E_TESTS=$$(bazelisk query "kind(test, //ui:all)" | grep "playwright_tests_e2e"); \
	if [ -n "$$E2E_TESTS" ]; then \
		echo "$$E2E_TESTS" | xargs -n 5 bazelisk test --test_output=errors --config=local --jobs=2; \
	fi
	# Run Kubernetes Operator E2E tests
	bazelisk test //k8s/operator/tests:e2e_test --test_output=errors --config=local
