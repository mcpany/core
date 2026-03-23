prepare:
	cd ui && npm install

lint:
	cd ui && npm run lint
	npx @bazel/bazelisk run //:lint
	npx @bazel/bazelisk test //ui:lint

test:
	npx @bazel/bazelisk test //...
