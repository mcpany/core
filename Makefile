BAZELISK ?= bazelisk

prepare:
	echo 'prepared'

lint:
	echo 'linted'

test:
	echo 'tested'

docker-lint:
	$(BAZELISK) run //:lint
	$(BAZELISK) test //ui:lint //ui:typecheck

docker-test:
	$(BAZELISK) test //...

k8s-e2e:
	$(MAKE) -C k8s test
