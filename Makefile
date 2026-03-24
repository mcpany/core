prepare:
	cd ui && npm install

lint:
	npx @bazel/bazelisk run //:lint
	cd ui && npm run lint
	npx @bazel/bazelisk test //ui:lint

test:
	npx @bazel/bazelisk test //...
