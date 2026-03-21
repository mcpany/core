prepare:
	npm install -g pre-commit
	cd ui && npm ci

lint:
	cd ui && npm run lint
	npx @bazel/bazelisk run //:lint

test:
	npx @bazel/bazelisk test //...
