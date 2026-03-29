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

run:
	bazelisk run //:mcpany -- run -c examples/hello_world.yaml
