prepare:
	sudo apt-get update && sudo apt-get install -y python3-pip python3-venv nodejs npm
	npm install -g @bazel/bazelisk
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/local/bin v1.64.6

lint:
	cd server && go test ./...
	bazelisk test //ui:lint //ui:typecheck

test:
	bazelisk test //server/...
