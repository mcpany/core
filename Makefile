.PHONY: prepare lint
prepare:
	mkdir -p build/env/bin

lint:
	cd server && go fmt ./pkg/util/file.go
