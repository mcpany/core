prepare:
	cd ui && npm install

lint:
	cd server && pre-commit run --files ../ui/src/components/ui/tag-input.tsx ../ui/tests/e2e/bulk_actions.spec.ts ../ui/src/components/services/service-list.tsx
	cd ui && npm run lint
	npx @bazel/bazelisk run //:lint
	npx @bazel/bazelisk test //ui:lint

test:
	npx @bazel/bazelisk test //...
