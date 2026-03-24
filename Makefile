prepare:
	sudo apt-get update && sudo apt-get install -y python3-pip python3-venv nodejs npm
	npm install -g @bazel/bazelisk
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.6

lint:
	export PATH=$(go env GOPATH)/bin:$$PATH && bazelisk run //:lint
	bazelisk test //ui:lint //ui:typecheck

test:
	bazelisk test //server/...
