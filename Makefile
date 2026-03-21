.PHONY: docker-lint docker-test k8s-e2e test lint

docker-lint:
	bazelisk run //:lint

docker-test:
	bazelisk test //server/... //proto/...

k8s-e2e:
	$(MAKE) -C k8s test

test: docker-test

lint: docker-lint
