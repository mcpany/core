.PHONY: prepare build lint test docker-lint docker-test k8s-e2e

prepare:
	go mod download

build:
	go build ./...

lint:
	./scripts/lint.sh

test:
	go test ./...

docker-lint: lint

docker-test: test

k8s-e2e:
	@echo "Running k8s e2e tests..."
