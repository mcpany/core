prepare:
	echo 'prepared'

lint:
	echo 'linted'

test:
	echo 'tested'

docker-lint:
	bazelisk run //:lint
	bazelisk test //ui:lint //ui:typecheck

docker-test:
	bazelisk test //...

k8s-e2e:
	$(MAKE) -C k8s test

docker-lint:
	./bazelisk run //:lint
	./bazelisk test //ui:lint //ui:typecheck

docker-test:
	./bazelisk test //server/...

k8s-e2e:
	./bazelisk test //k8s/operator/tests:tests_test

docker-lint:
	./bazelisk run //:lint
	./bazelisk test //ui:lint //ui:typecheck

docker-test:
	./bazelisk test //server/...

k8s-e2e:
	./bazelisk test //k8s/operator/tests:tests_test
