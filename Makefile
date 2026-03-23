prepare:
	sudo apt-get install -y python3-pip python3-venv
	python3 -m venv venv
	venv/bin/pip install pre-commit
	cd ui && npm ci

lint:
	venv/bin/pre-commit run --all-files
	cd ui && npm run lint
	npx @bazel/bazelisk run //:lint
	npx @bazel/bazelisk test //ui:lint

test:
	npx @bazel/bazelisk test //...
