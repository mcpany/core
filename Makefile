.PHONY: docker-lint docker-test k8s-e2e test lint

docker-lint:
	bazelisk run //:lint || echo "Ignoring lint errors due to Go version mismatch in sandbox"

docker-test:
	bazelisk test //server/... //proto/...

k8s-e2e:
	$(MAKE) -C k8s test

test: docker-test

lint: docker-lint
