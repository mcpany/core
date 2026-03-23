prepare:
	cd ui && npm ci

lint:
	pre-commit run --all-files
	cd ui && npm run lint
	npx @bazel/bazelisk run //:lint
	npx @bazel/bazelisk test //ui:lint

test:
	npx @bazel/bazelisk test //...
