# Shim Makefile to forward commands to server/Makefile and ui/Makefile
.PHONY: all test lint build run clean gen

# Targets that should run on both (or just server generally, but test on both)
all: gen
	$(MAKE) -C server all
	$(MAKE) -C ui build

test: gen
	$(MAKE) -C server test
	$(MAKE) -C ui test

k8s-e2e:
	./tests/k8s/test_operator.sh

lint:
	$(MAKE) -C server lint

# Run runs server
run:
	$(MAKE) -C server run

clean:
	$(MAKE) -C server clean
	# ui clean if needed, likely just removing node_modules or build artifacts?

gen: gen-go gen-ts

gen-go:
	$(MAKE) -C server gen

gen-ts:
	@echo "Generating Protobuf TypeScript definitions..."
	@npm install -g ts-proto
	@./build/env/bin/protoc \
	--plugin="protoc-gen-ts_proto=$(shell which protoc-gen-ts_proto)" \
	--ts_proto_out=. \
	--ts_proto_opt=esModuleInterop=true,outputServices=nice-grpc,outputJsonMethods=true,env=browser \
	--proto_path=. \
	proto/config/v1/config.proto \
	proto/config/v1/upstream_service.proto \
	proto/config/v1/auth.proto \
	proto/config/v1/tool.proto \
	proto/config/v1/resource.proto \
	proto/config/v1/prompt.proto \
	proto/config/v1/health_check.proto \
	proto/config/v1/webhook.proto
	@mkdir -p proto/google
	@if [ -d "google/protobuf" ]; then \
		rm -rf proto/google/protobuf; \
		mv google/protobuf proto/google/; \
		rm -rf google; \
	fi
	@find proto -name "*.ts" -exec sed -i 's|\.\./\.\./\.\./google|\.\./\.\./google|g' {} +

%:
	$(MAKE) -C server $@
