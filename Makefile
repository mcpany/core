# Shim Makefile to forward commands to server/Makefile and ui/Makefile
.PHONY: all test lint build run clean gen prepare-proto clean-protos

# Variables
GO = go
GO_CMD := $(GO)
BUILD_DIR := $(abspath ./build)
TOOL_INSTALL_DIR := $(BUILD_DIR)/env/bin
PROTOC_INCLUDE_DIR := $(TOOL_INSTALL_DIR)/include
GOOGLEAPIS_DIR := $(BUILD_DIR)/googleapis

# Variables for protoc installation
PROTOC_VERSION_URL := https://api.github.com/repos/protocolbuffers/protobuf/releases/latest
PROTOC_DOWNLOAD_URL_BASE := https://github.com/protocolbuffers/protobuf/releases/download
PROTOC_GEN_GO_VERSION ?= v1.36.11
PROTOC_GEN_GO_GRPC_VERSION ?= v1.5.1
GRPC_GATEWAY_VERSION ?= v2.27.3
PROTOC_ZIP := protoc.zip
PROTOC_VERSION := v33.1

# Detect architecture for protoc
UNAME_M := $(shell uname -m)
PROTOC_ARCH := $(UNAME_M)
ifeq ($(UNAME_M), aarch64)
	PROTOC_ARCH := aarch_64
endif

PROTOC_GEN_GO := $(TOOL_INSTALL_DIR)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(TOOL_INSTALL_DIR)/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(TOOL_INSTALL_DIR)/protoc-gen-grpc-gateway
PROTOC_GEN_OPENAPIV2 := $(TOOL_INSTALL_DIR)/protoc-gen-openapiv2
PROTOC_BIN := $(TOOL_INSTALL_DIR)/protoc

# Targets that should run on both (or just server generally, but test on both)
all: gen
	$(MAKE) -C server all
	$(MAKE) -C ui build

prepare:
	$(MAKE) -C server prepare

test: gen
	$(MAKE) test-proto
	$(MAKE) -C server test
	$(MAKE) -C ui test
	$(MAKE) -C k8s test


docker-build-all:
	$(MAKE) -C server docker-build-server docker-build-dev docker-build-http-echo
	$(MAKE) -C ui docker-build-ui build-test-docker
test-proto:
	@echo "Running proto tests..."
	@go test ./proto/...

k8s-e2e:
	$(MAKE) -C k8s test

k8s-test: k8s-e2e

lint:
	$(MAKE) -C server lint

# Run runs server
run:
	$(MAKE) -C server run

clean:
	$(MAKE) -C server clean
	# ui clean if needed, likely just removing node_modules or build artifacts?

prepare-proto:
	@echo "Preparing protobuf environment..."
	@mkdir -p $(TOOL_INSTALL_DIR)
	@# Check if protoc is installed
	@export PATH=$(TOOL_INSTALL_DIR):$$PATH; \
	PROTOC_TAG=$(PROTOC_VERSION); \
	if test -f "$(PROTOC_BIN)"; then \
		INSTALLED_VERSION=v$$($(PROTOC_BIN) --version | sed 's/libprotoc //'); \
		if test "$${INSTALLED_VERSION}" = "$${PROTOC_TAG}"; then \
			echo "protoc version $${INSTALLED_VERSION} is already installed."; \
		else \
			echo "protoc version mismatch. Installed: $${INSTALLED_VERSION}, Required: $${PROTOC_TAG}. Re-installing..."; \
			rm -f "$(PROTOC_BIN)"; \
			$(MAKE) prepare-proto; \
		fi; \
	else \
		echo "protoc not found, attempting to install version $${PROTOC_TAG}..."; \
		if ! command -v curl >/dev/null 2>&1 || ! command -v unzip >/dev/null 2>&1; then \
			echo "curl and unzip are not installed. Installing..."; \
			apt-get update && apt-get install -y curl unzip; \
		fi; \
		PROTOC_VERSION_NO_V=$$(echo "$${PROTOC_TAG}" | sed 's/v//'); \
		PROTOC_DOWNLOAD_URL_NO_V="$(PROTOC_DOWNLOAD_URL_BASE)/$${PROTOC_TAG}/protoc-$${PROTOC_VERSION_NO_V}-linux-$(PROTOC_ARCH).zip"; \
		echo "Downloading protoc from $${PROTOC_DOWNLOAD_URL_NO_V}..."; \
		if curl -sSL "$${PROTOC_DOWNLOAD_URL_NO_V}" -o "$(PROTOC_ZIP)"; then \
			echo "Unzipping to $(TOOL_INSTALL_DIR)..."; \
			unzip -o "$(PROTOC_ZIP)" -d "$(TOOL_INSTALL_DIR)"; \
			mv "$(TOOL_INSTALL_DIR)/bin/protoc" "$(PROTOC_BIN)"; \
			rm -f "$(PROTOC_ZIP)"; \
		fi; \
	fi
	@# Install Go protobuf plugins
	@echo "Installing Go protobuf plugins..."
	@if ! test -f "$(PROTOC_GEN_GO)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION); fi
	@if ! test -f "$(PROTOC_GEN_GO_GRPC)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION); fi
	@if ! test -f "$(PROTOC_GEN_GRPC_GATEWAY)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GRPC_GATEWAY_VERSION); fi
	@if ! test -f "$(PROTOC_GEN_OPENAPIV2)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(GRPC_GATEWAY_VERSION); fi
	@# Download grpc-gateway source for protos
	@if ! test -d "$(BUILD_DIR)/grpc-gateway"; then \
		echo "Downloading grpc-gateway protos..."; \
		curl -sSL -o grpc-gateway.zip https://github.com/grpc-ecosystem/grpc-gateway/archive/refs/tags/$(GRPC_GATEWAY_VERSION).zip; \
		unzip -q grpc-gateway.zip -d $(BUILD_DIR); \
		GRPC_GATEWAY_VER_NO_V=$$(echo "$(GRPC_GATEWAY_VERSION)" | sed 's/v//'); \
		mv $(BUILD_DIR)/grpc-gateway-$$GRPC_GATEWAY_VER_NO_V $(BUILD_DIR)/grpc-gateway; \
		rm grpc-gateway.zip; \
	fi
	@# Download googleapis
	@if ! test -d "$(BUILD_DIR)/googleapis"; then \
		echo "Downloading googleapis..."; \
		curl -sSL -o googleapis.zip https://github.com/googleapis/googleapis/archive/refs/heads/master.zip; \
		unzip -q googleapis.zip -d $(BUILD_DIR); \
		mv $(BUILD_DIR)/googleapis-master $(BUILD_DIR)/googleapis; \
		rm googleapis.zip; \
	fi

clean-protos:
	@echo "Cleaning generated protobuf files..."
	@-find proto -name "*.ts" -delete
	@-rm -rf proto/google
	@-find proto server/pkg server/cmd -name "*.pb.go" -delete
	@-find proto -name "*.pb.gw.go" -delete

gen: clean-protos prepare-proto
	@echo "Generating protobuf files (Go)..."
	@export PATH=$(TOOL_INSTALL_DIR):$$PATH; \
		mkdir -p $(BUILD_DIR); \
		find proto -name "*.proto" -not -path "proto/third_party/*" -not -path "proto/google/*" -exec protoc \
			--proto_path=. \
			--proto_path=$(BUILD_DIR)/grpc-gateway \
			--proto_path=$(BUILD_DIR)/googleapis \
			--descriptor_set_out=$(BUILD_DIR)/all.protoset \
			--include_imports \
			--go_out=. \
			--go_opt=module=github.com/mcpany/core,default_api_level=API_HYBRID \
			--go-grpc_out=. \
			--go-grpc_opt=module=github.com/mcpany/core \
			--grpc-gateway_out=. \
			--grpc-gateway_opt=module=github.com/mcpany/core \
			{} +; \
		rm -rf google


	@echo "Generating protobuf files (TypeScript)..."
	@if ! [ -f "./ui/node_modules/.bin/protoc-gen-ts_proto" ]; then \
		echo "protoc-gen-ts_proto not found. Installing UI dependencies..."; \
		cd ui && (npm ci || npm install || (sleep 5 && npm install)); \
	fi
	@if [ -f "./ui/node_modules/.bin/protoc-gen-ts_proto" ]; then \
		export PATH=$(TOOL_INSTALL_DIR):$$PATH; \
		find proto -name "*.proto" -exec protoc \
			--proto_path=. \
			--proto_path=$(BUILD_DIR)/grpc-gateway \
			--proto_path=$(GOOGLEAPIS_DIR) \
			--plugin=protoc-gen-ts_proto=./ui/node_modules/.bin/protoc-gen-ts_proto \
			--ts_proto_out=. \
			--ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
			{} +; \
		if [ -d "google" ]; then mv google proto/; fi; \
		if [ -f "proto/google/protobuf/struct.ts" ]; then sed -i 's/map((e) => e)/map((e: any) => e)/g' proto/google/protobuf/struct.ts; fi; \
		find proto -name "*.ts" -exec sed -i 's|\.\./\.\./\.\./google|\.\./\.\./google|g' {} +; \
		echo "Local TypeScript Protobuf generation complete."; \
		echo "Generating standard protobuf files (TypeScript)..."; \
		STANDARD_PROTOS=$$(find $(PROTOC_INCLUDE_DIR)/google/protobuf $(GOOGLEAPIS_DIR)/google/api -name "*.proto" 2>/dev/null); \
		if [ -n "$$STANDARD_PROTOS" ]; then \
			protoc \
				--proto_path=$(PROTOC_INCLUDE_DIR) \
				--proto_path=$(GOOGLEAPIS_DIR) \
				--plugin=protoc-gen-ts_proto=./ui/node_modules/.bin/protoc-gen-ts_proto \
				--ts_proto_out=proto \
				--ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
				$$STANDARD_PROTOS; \
		fi; \
		echo "Standard TypeScript Protobuf generation complete."; \
	else \
		echo "Warning: protoc-gen-ts_proto not found in ./ui/node_modules/.bin/. Skipping TypeScript generation."; \
	fi

update-screenshots:
	$(MAKE) -C ui update-screenshots

# Forward other targets to server by default
%:
	$(MAKE) -C server $@
