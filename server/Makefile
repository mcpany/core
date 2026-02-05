
# Makefile

# Variables
GO = go
GO_CMD := $(GO)
ifdef INSIDE_DOCKER_CONTAINER
GO_CMD := /usr/local/go/bin/go
endif
SERVER_IMAGE_TAGS ?= mcpany/server:latest
EVERYTHING_IMAGE_TAG ?= mcpany/everything:latest
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags="-X main.Version=$(VERSION)"
BUILD_DIR := $(abspath ../build)
# Use /tmp for bundles to avoid Docker bind mount permission issues in some environments
MCP_BUNDLE_DIR ?= $(BUILD_DIR)/bundles

# Skip Docker tests in CI to avoid overlayfs issues
ifneq ($(CI),)
SKIP_DOCKER_TESTS := true
endif

# Dockerized Build Logic
BUILD_MODE ?= local
DOCKER_COMPOSE_REGULAR_FILE ?= docker/docker-compose.dev.yml
ifeq ($(BUILD_MODE),docker)
ifneq ($(INSIDE_DOCKER_CONTAINER),1)
	SHOULD_PROXY_TO_DOCKER := true
	DOCKER_GID := $(shell getent group docker | cut -d: -f3)
	DOCKER_PROXY_CMD := env UID=$(shell id -u) GID=$(shell id -g) DOCKER_GID=$(DOCKER_GID) docker compose -f $(DOCKER_COMPOSE_REGULAR_FILE) run --rm dev make INSIDE_DOCKER_CONTAINER=1
endif
endif

HAS_DOCKER := $(shell command -v docker 2> /dev/null)
# Check if docker can be run without sudo
ifeq ($(shell docker info >/dev/null 2>&1; echo $$?), 0)
	DOCKER_CMD := docker
	DOCKER_BUILDX_CMD := docker buildx
	ifeq ($(shell docker buildx version >/dev/null 2>&1 && echo 1 || echo 0), 1)
	DOCKER_BUILD_CMD := DOCKER_API_VERSION=1.44 docker buildx build --load
	HAS_BUILDX := 1
else
	DOCKER_BUILD_CMD := docker build
	HAS_BUILDX := 0
endif
	SUDO_MSG := " (no sudo)"
	NEEDS_SUDO_FOR_DOCKER := 0
else
	DOCKER_CMD := sudo docker
	DOCKER_BUILDX_CMD := sudo docker buildx
	ifeq ($(shell sudo docker buildx version >/dev/null 2>&1 && echo 1 || echo 0), 1)
		DOCKER_BUILD_CMD := sudo docker buildx build --load
		HAS_BUILDX := 1
	else
		DOCKER_BUILD_CMD := sudo docker build
		HAS_BUILDX := 0
	endif
	NEEDS_SUDO_FOR_DOCKER := 1
endif

# Variables for protoc installation
PROTOC_VERSION_URL := https://api.github.com/repos/protocolbuffers/protobuf/releases/latest
PROTOC_DOWNLOAD_URL_BASE := https://github.com/protocolbuffers/protobuf/releases/download
PROTOC_GEN_GO_VERSION ?= v1.36.11
PROTOC_GEN_GO_GRPC_VERSION ?= v1.5.1
GRPC_GATEWAY_VERSION ?= v2.27.3
PROTOC_ZIP := protoc.zip
# Debug TOOL_INSTALL_DIR

ifdef INSIDE_DOCKER_CONTAINER
TOOL_INSTALL_DIR := /tools
endif
TOOL_INSTALL_DIR ?= $(BUILD_DIR)/env/bin
PROTOC_VERSION := v33.1

# Detect architecture for protoc
UNAME_M := $(shell uname -m)
PROTOC_ARCH := $(UNAME_M)
ifeq ($(UNAME_M), aarch64)
	PROTOC_ARCH := aarch_64
endif

LOCAL_BIN_DIR := $(BUILD_DIR)/bin
PROTOC_GEN_GO := $(TOOL_INSTALL_DIR)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(TOOL_INSTALL_DIR)/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(TOOL_INSTALL_DIR)/protoc-gen-grpc-gateway
PROTOC_GEN_OPENAPIV2 := $(TOOL_INSTALL_DIR)/protoc-gen-openapiv2
PROTOC_BIN := $(TOOL_INSTALL_DIR)/protoc
GOLANGCI_LINT_BIN := $(TOOL_INSTALL_DIR)/golangci-lint
GOFUMPT_BIN := $(TOOL_INSTALL_DIR)/gofumpt
GOIMPORTS_BIN := $(TOOL_INSTALL_DIR)/goimports
PRE_COMMIT_VERSION := 4.3.0
PRE_COMMIT_BIN := $(TOOL_INSTALL_DIR)/pre-commit
HELM_BIN := $(TOOL_INSTALL_DIR)/helm
SHELLCHECK_BIN := $(TOOL_INSTALL_DIR)/shellcheck
NATS_SERVER_BIN := $(TOOL_INSTALL_DIR)/nats-server

# Helm installation variables
HELM_VERSION ?= v3.19.0
HELM_DOWNLOAD_URL_BASE ?= https://get.helm.sh

# Python executable determination
# In a CI environment, we expect Python to be set up and available in the PATH.
# Locally, we use a virtual environment.
ifeq ($(CI),)
	PYTHON_EXEC := $(BUILD_DIR)/venv/bin/python
else
	PYTHON_EXEC := python3
endif
# ==============================================================================
# Release Targets
# ==============================================================================
RELEASE_DIR := $(BUILD_DIR)/release
# PLATFORMS variable will be used to define the target platforms for the build.
# Example: PLATFORMS := linux/amd64 linux/arm64
# PLATFORMS ?= linux/amd64 linux/arm64

# Find all .proto files, excluding vendor/cache directories
PROTO_FILES := $(shell find ../proto -path ./vendor -prune -o -name "*.proto" -print)

.PHONY: all gen build test e2e clean run build-docker run-docker gen build test e2e-local check-local release release-local release-docker

all: build

# ==============================================================================
# Main Targets (Default to local, use USE_DOCKER=1 to switch to Docker)
# ==============================================================================


# ==============================================================================
# Local Commands
# ==============================================================================

prepare:
	@echo "Preparing development environment..."
	@mkdir -p $(LOCAL_BIN_DIR)
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
			$(MAKE) prepare; \
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
			if test -f "$(PROTOC_BIN)"; then \
				export PATH=$(TOOL_INSTALL_DIR):$$PATH; \
				echo "protoc installed successfully to $(TOOL_INSTALL_DIR). This directory has been added to PATH for this session."; \
				echo "You may want to add it to your system PATH permanently: export PATH=$(TOOL_INSTALL_DIR):$$PATH"; \
				$(PROTOC_BIN) --version; \
			else \
				echo "Error: protoc binary not found in $(TOOL_INSTALL_DIR) after unzip. The downloaded archive may not have the expected structure."; \
				exit 1; \
			fi; \
			rm -f "$(PROTOC_ZIP)"; \
		fi; \
	fi
	@# Install Go protobuf plugins
	@echo "Installing Go protobuf plugins..."
	@if test -f "$(PROTOC_GEN_GO)"; then \
		INSTALLED_VER=$$($(PROTOC_GEN_GO) --version | awk '{print $$2}'); \
		if [ "$$INSTALLED_VER" != "$(PROTOC_GEN_GO_VERSION)" ]; then \
			echo "protoc-gen-go version mismatch. Installed: $$INSTALLED_VER, Required: $(PROTOC_GEN_GO_VERSION). Re-installing..."; \
			rm -f "$(PROTOC_GEN_GO)"; \
		fi; \
	fi
	@if test -f "$(PROTOC_GEN_GO_GRPC)"; then \
		INSTALLED_VER=$$($(PROTOC_GEN_GO_GRPC) --version | awk '{print $$2}'); \
		if [ "$$INSTALLED_VER" != "$(PROTOC_GEN_GO_GRPC_VERSION)" ] && [ "v$$INSTALLED_VER" != "$(PROTOC_GEN_GO_GRPC_VERSION)" ]; then \
			echo "protoc-gen-go-grpc version mismatch. Installed: $$INSTALLED_VER, Required: $(PROTOC_GEN_GO_GRPC_VERSION). Re-installing..."; \
			rm -f "$(PROTOC_GEN_GO_GRPC)"; \
		fi; \
	fi
	@if ! test -f "$(PROTOC_GEN_GO)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION); fi
	@if ! test -f "$(PROTOC_GEN_GO_GRPC)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION); fi
	@if ! test -f "$(PROTOC_GEN_GRPC_GATEWAY)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GRPC_GATEWAY_VERSION); fi
	@if ! test -f "$(PROTOC_GEN_OPENAPIV2)"; then GOBIN=$(TOOL_INSTALL_DIR) $(GO_CMD) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(GRPC_GATEWAY_VERSION); fi
	@echo "Checking for Go protobuf plugins..."
	@if ! test -f "$(PROTOC_GEN_GO)"; then \
		echo "protoc-gen-go not found at $(PROTOC_GEN_GO) after attempting install. Please check your GOPATH/GOBIN setup and PATH."; \
		exit 1; \
	fi
	@if ! test -f "$(PROTOC_GEN_GO_GRPC)"; then \
		echo "protoc-gen-go-grpc not found at $(PROTOC_GEN_GO_GRPC) after attempting install. Please check your GOPATH/GOBIN setup and PATH."; \
		exit 1; \
	fi
	@if ! test -f "$(PROTOC_GEN_GRPC_GATEWAY)"; then \
		echo "protoc-gen-grpc-gateway not found at $(PROTOC_GEN_GRPC_GATEWAY) after attempting install."; \
		exit 1; \
	fi
	@echo "Go protobuf plugins installation check complete."
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
	@# Install Python dependencies and pre-commit hooks
	@echo "Checking for Python to install dependencies and pre-commit hooks..."
	@if command -v python3 >/dev/null 2>&1; then \
		PYTHON_CMD=python3; \
	elif command -v python >/dev/null 2>&1; then \
		PYTHON_CMD=python; \
	else \
		PYTHON_CMD=""; \
	fi; \
	if test -z "$(CI)"; then \
		if test -n "$$PYTHON_CMD"; then \
			echo "Python found. Installing/updating pre-commit and fastmcp locally..."; \
			VENV_DIR=$(BUILD_DIR)/venv; \
			if ! test -d "$$VENV_DIR"; then \
				echo "Creating virtual environment..."; \
				$$PYTHON_CMD -m venv $$VENV_DIR --without-pip; \
				curl -sSL https://bootstrap.pypa.io/get-pip.py -o $$VENV_DIR/get-pip.py; \
				$$VENV_DIR/bin/python $$VENV_DIR/get-pip.py; \
			fi; \
			if ! test -f "$$VENV_DIR/bin/python"; then \
				echo "\n\033[1;31mERROR: Failed to create Python virtual environment.\033[0m"; \
				exit 1; \
			fi; \
			if ! test -f "$$VENV_DIR/.installed"; then \
				$(PYTHON_EXEC) -m pip install --upgrade pip; \
				$(PYTHON_EXEC) -m pip install -r requirements.txt; \
				touch "$$VENV_DIR/.installed"; \
			fi; \
			mkdir -p $(BUILD_DIR)/.cache/pre-commit; \
			if ! GIT_CONFIG_GLOBAL=/dev/null PRE_COMMIT_HOME=$(BUILD_DIR)/.cache/pre-commit $(PYTHON_EXEC) -m pre_commit install; then \
				echo "\n\033[1;33mWARNING: pre-commit hook installation failed.\033[0m"; \
				echo "\033[1;33mThis is likely because a global git hooks path is configured (core.hooksPath).\033[0m"; \
				echo "\033[1;33mThe build will continue, but pre-commit hooks will not be active.\033[0m"; \
				echo "\033[1;33mTo fix this, run: git config --unset-all core.hooksPath\n"; \
			fi; \
		else \
			echo "Python not found, skipping Python dependency installation and pre-commit hook setup."; \
		fi; \
	else \
		echo "CI environment detected, skipping local Python venv setup."; \
	fi
	@#
	@echo "Installing Node.js dependencies for integration tests..."
	@if command -v npm >/dev/null 2>&1; then \
		if test -f "tests/integration/upstream/package.json"; then \
			echo "Found package.json, running npm install..."; \
			(cd tests/integration/upstream && npm install) || true; \
			if [ -n "$(CI)" ]; then \
				echo "CI detected, installing Playwright with system dependencies..."; \
				(cd tests/integration/upstream && npx playwright install --with-deps) || true; \
			else \
				echo "Installing Playwright browsers (skipping system deps)..."; \
				(cd tests/integration/upstream && npx playwright install) || true; \
			fi; \
		else \
			echo "No package.json found in tests/integration/upstream, skipping npm install."; \
		fi; \
	else \
		echo "npm not found, skipping Node.js dependency installation."; \
	fi
	@# Install Helm
	@echo "Checking for Helm..."
	@if test -f "$(HELM_BIN)"; then \
		echo "Helm is already installed."; \
	else \
		echo "Installing Helm to $(TOOL_INSTALL_DIR)..."; \
		OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		ARCH=$$(uname -m); \
		case $${ARCH} in \
			x86_64) ARCH=amd64 ;; \
			aarch64) ARCH=arm64 ;; \
		esac; \
		HELM_TARBALL="helm-$(HELM_VERSION)-$${OS}-$${ARCH}.tar.gz"; \
		HELM_DOWNLOAD_URL="$(HELM_DOWNLOAD_URL_BASE)/$${HELM_TARBALL}"; \
		echo "Downloading Helm from $${HELM_DOWNLOAD_URL}..."; \
		if curl -sSL "$${HELM_DOWNLOAD_URL}" -o "/tmp/$${HELM_TARBALL}"; then \
			echo "Unpacking to $(TOOL_INSTALL_DIR)..."; \
			tar -zxvf "/tmp/$${HELM_TARBALL}" -C "/tmp"; \
			mv "/tmp/$${OS}-$${ARCH}/helm" "$(HELM_BIN)"; \
			rm "/tmp/$${HELM_TARBALL}"; \
			rm -rf "/tmp/$${OS}-$${ARCH}"; \
			if test -f "$(HELM_BIN)"; then \
				echo "Helm installed successfully."; \
			else \
				echo "Helm installation failed."; \
				exit 1; \
			fi; \
		else \
			echo "Failed to download Helm."; \
			exit 1; \
		fi; \
	fi
	@# Install ShellCheck
	@echo "Checking for ShellCheck..."
	@if test -f "$(SHELLCHECK_BIN)"; then \
		echo "ShellCheck is already installed."; \
	else \
		echo "Installing ShellCheck to $(TOOL_INSTALL_DIR)..."; \
		SC_VERSION="stable"; \
		wget -qO- "https://github.com/koalaman/shellcheck/releases/download/$${SC_VERSION}/shellcheck-$${SC_VERSION}.linux.x86_64.tar.xz" | tar -xJv -C /tmp; \
		mv "/tmp/shellcheck-$${SC_VERSION}/shellcheck" "$(SHELLCHECK_BIN)"; \
		rm -rf "/tmp/shellcheck-$${SC_VERSION}"; \
		if test -f "$(SHELLCHECK_BIN)"; then \
			echo "ShellCheck installed successfully."; \
		else \
			echo "ShellCheck installation failed."; \
			exit 1; \
		fi; \
	fi
	@# Install nats-server
	@echo "Checking for nats-server..."
	@if test -f "$(NATS_SERVER_BIN)"; then \
		echo "nats-server is already installed."; \
	else \
		echo "Installing nats-server to $(TOOL_INSTALL_DIR)..."; \
		GOBIN=$(TOOL_INSTALL_DIR) go install github.com/nats-io/nats-server/v2@latest; \
		if test -f "$(NATS_SERVER_BIN)"; then \
			echo "nats-server installed successfully."; \
		else \
			echo "nats-server installation failed."; \
			exit 1; \
		fi; \
	fi
	@# Install golangci-lint
	@echo "Checking for golangci-lint..."
	@LINT_VER_CHECK=$$($(GOLANGCI_LINT_BIN) --version 2>/dev/null | grep "v2.7.2"); \
	if [ -n "$$LINT_VER_CHECK" ]; then \
		echo "golangci-lint v2.7.2 is already installed."; \
	else \
		echo "Installing golangci-lint v2.7.2 to $(TOOL_INSTALL_DIR)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOL_INSTALL_DIR) v2.7.2; \
		if test -f "$(GOLANGCI_LINT_BIN)"; then \
			echo "golangci-lint installed successfully."; \
		else \
			echo "golangci-lint installation failed."; \
			exit 1; \
		fi; \
	fi
	@echo "Downloading go modules..." && \
	go mod download
	@echo "Preparation complete."




# Delegate gen to root Makefile
gen: prepare
	$(MAKE) -C .. gen

build: gen
	@echo "Building Go project locally..."
	@$(GO_CMD) build $(LDFLAGS) -buildvcs=false -o $(BUILD_DIR)/bin/server ./cmd/server

build-mock-server:
	@echo "Building mock server..."
	@$(GO_CMD) build $(LDFLAGS) -buildvcs=false -o $(BUILD_DIR)/bin/mock_server ./cmd/mock_mcp_server

ifdef SHOULD_PROXY_TO_DOCKER
test-server:
	@$(DOCKER_PROXY_CMD) test
else
test-server: test-fast e2e-local test-public-api
endif

test: test-server

COVERAGE_FILE ?= coverage.out

e2e-local: e2e-parallel e2e-sequential

e2e-parallel: build-test-images build build-examples build-e2e-mocks
	@echo "Running parallel E2E Go tests locally with a 1200s timeout..."
	@-pkill -f file-upload-server || true
	@$(EXAMPLE_BIN_DIR)/file-upload-server &
	@mkdir -p $(BUILD_DIR)/bundles
	@GEMINI_API_KEY=$(GEMINI_API_KEY) MCP_BUNDLE_DIR=$(MCP_BUNDLE_DIR) MCPANY_DEBUG=true MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true CGO_ENABLED=1 USE_SUDO_FOR_DOCKER=$(NEEDS_SUDO_FOR_DOCKER) SHARD=$(SHARD) MCPANY_PROJECT_ROOT=$(CURDIR) ./tests/run_shard.sh -race -count=1 -timeout 1200s -tags=e2e -cover -coverprofile=$(BUILD_DIR)/coverage.e2e-parallel.out
	@-pkill -f file-upload-server

e2e-sequential: build-test-images build build-examples build-e2e-mocks
	@echo "Running sequential E2E Go tests locally with a 300s timeout..."
	@GEMINI_API_KEY=$(GEMINI_API_KEY) MCP_BUNDLE_DIR=$(MCP_BUNDLE_DIR) MCPANY_DEBUG=true MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true CGO_ENABLED=1 USE_SUDO_FOR_DOCKER=$(NEEDS_SUDO_FOR_DOCKER) E2E_DOCKER=true MCPANY_PROJECT_ROOT=$(CURDIR) $(GO_CMD) test -p 1 -parallel 1 -race -count=1 -timeout 600s -tags=e2e -cover -coverprofile=$(BUILD_DIR)/coverage.e2e-sequential.out ./tests/e2e_sequential/...

test-fast: build build-examples build-e2e-mocks build-e2e-timeserver-docker
	@echo "Running fast Go tests locally with a 300s timeout..."
	@TEST_PKGS=$$(go list ./... | grep -v /tests/e2e | grep -v /tests/e2e_sequential | grep -v /tests/integration | grep -v /tests/public_api | grep -v /build | grep -v /vendor | grep -v /proto); \
	echo "Testing packages: $$TEST_PKGS"; \
	if [ -z "$$TEST_PKGS" ]; then echo "No packages to test!"; exit 1; fi; \
	mkdir -p $(BUILD_DIR)/bundles; \
	GEMINI_API_KEY=$(GEMINI_API_KEY) MCP_BUNDLE_DIR=$(MCP_BUNDLE_DIR) MCPANY_DEBUG=true MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true CGO_ENABLED=1 USE_SUDO_FOR_DOCKER=$(NEEDS_SUDO_FOR_DOCKER) $(GO_CMD) test -race -count=1 -timeout 600s -cover -coverprofile=$(COVERAGE_FILE) $$TEST_PKGS

.PHONY: test-public-api
test-public-api: build build-mock-server
	@echo "Running public API E2E Go tests with a 300s timeout..."
	@GEMINI_API_KEY=$(GEMINI_API_KEY) MCPANY_DEBUG=true MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true CGO_ENABLED=1 USE_SUDO_FOR_DOCKER=$(NEEDS_SUDO_FOR_DOCKER) $(GO_CMD) test -parallel 1 -race -count=1 -timeout 300s -tags=e2e_public_api -cover -coverprofile=$(COVERAGE_FILE) ./tests/public_api/...

# ==============================================================================
# Example Binaries Build
# ==============================================================================
EXAMPLE_BIN_DIR := $(BUILD_DIR)/examples/bin

# List of example binaries
.PHONY: build-examples
build-examples: build-weather-server build-stdio-tool

# Target to build the weather server example
.PHONY: build-weather-server
build-weather-server:
	@echo "Building weather server example..."
	@mkdir -p $(EXAMPLE_BIN_DIR)
	@$(GO_CMD) build -buildvcs=false -o $(EXAMPLE_BIN_DIR)/weather-server examples/upstream_service_demo/http/server/weather_server/weather_server.go

# Target to build the stdio tool example
.PHONY: build-stdio-tool
build-stdio-tool:
	@echo "Building stdio tool example..."
	@$(GO_CMD) build -buildvcs=false -o examples/demo/stdio/my-tool-bin examples/demo/stdio/my-tool/main.go

.PHONY: build-file-upload-server
build-file-upload-server:
	@echo "Building file upload server example..."
	@mkdir -p $(EXAMPLE_BIN_DIR)
	@$(GO_CMD) build -buildvcs=false -o $(EXAMPLE_BIN_DIR)/file-upload-server examples/upstream_service_demo/file-upload/main.go

# UI targets removed (moved to ui/Makefile)

.PHONY: e2e
e2e: e2e-local



lint:
ifdef SHOULD_PROXY_TO_DOCKER
	@$(DOCKER_PROXY_CMD) lint
else
	@echo "Running lint (using pre-commit)..."
	@mkdir -p $(BUILD_DIR)/.cache/pre-commit $(BUILD_DIR)/.cache/golangci-lint
	@export PATH=$(TOOL_INSTALL_DIR):$(BUILD_DIR)/venv/bin:$$PATH && \
	export PRE_COMMIT_HOME=$(BUILD_DIR)/.cache/pre-commit && \
	export GOLANGCI_LINT_CACHE=$(BUILD_DIR)/.cache/golangci-lint && \
	export GOWORK=off && \
	$(GOLANGCI_LINT_BIN) run --timeout 10m --fix ./cmd/... ./pkg/... ./tests/... ./examples/... && \
	cd .. && pre-commit run --config server/.pre-commit-config.yaml --all-files
endif

clean-pre-commit:
ifdef SHOULD_PROXY_TO_DOCKER
	@$(DOCKER_PROXY_CMD) clean-pre-commit
else
	@export PATH=$(TOOL_INSTALL_DIR):$$PATH && \
	export PRE_COMMIT_HOME=$(BUILD_DIR)/.cache/pre-commit && \
	pre-commit clean
endif

.PHONY: fix-license-header
fix-license-header:
	@echo "Fixing license headers..."
	@echo "Step 1: Removing existing headers..."
	@$(GO_CMD) run tools/license/remove.go
	@echo "Step 2: Adding correct headers via pre-commit..."
	@export GOWORK=off && pre-commit run addlicense --all-files || true
.PHONY: clean-protos
clean-protos:
	@echo "Cleaning generated protobuf files..."
	@-find . \( -name ".cache" -o -name "build" \) -prune -o -name "*.pb.go" -type f -exec rm -f {} +

clean: clean-protos
	@echo "Cleaning build artifacts..."
	@-chmod -R +w $(BUILD_DIR) 2>/dev/null || true
	@rm -rf $(BUILD_DIR)

run: build
	@echo "Starting MCP Any server locally..."
	@$(BUILD_DIR)/bin/server run --config-path=$(abspath config.minimal.yaml)

# ==============================================================================
# E2E Test Related Builds
# ==============================================================================
E2E_MOCK_DIR := $(CURDIR)/tests/integration/cmd/mocks
E2E_BIN_DIR := $(BUILD_DIR)/test/bin

# List of mock service directories (which are also their binary names)
E2E_MOCK_SERVICES := http_echo_server http_authed_echo_server grpc_weather_server grpc_authed_weather_server openapi_weather_server webrtc_weather_server command-tester prompt-server

# Target to build all E2E mock services
.PHONY: build-e2e-mocks
build-e2e-mocks:
	@mkdir -p $(E2E_BIN_DIR)
	@$(MAKE) $(addprefix $(E2E_BIN_DIR)/,$(E2E_MOCK_SERVICES))
	@$(MAKE) build-file-upload-server

# Rule to build a single E2E mock service
# < is the first prerequisite (the main.go file)
# * is the stem of the pattern match (the service name)
# @ is the target name (the output binary path)
$(E2E_BIN_DIR)/%: $(E2E_MOCK_DIR)/%/main.go
	@echo "Building E2E mock service: $* from $< into $(E2E_BIN_DIR)"
	@$(GO_CMD) build -buildvcs=false -o $@ $<

$(E2E_BIN_DIR)/command-tester: tests/integration/cmd/command-tester/main.go
	@echo "Building E2E mock service: command-tester from $< into $(E2E_BIN_DIR)"
	@$(GO_CMD) build -buildvcs=false -o $@ $<

# Target to ensure the E2E binary directory exists
$(E2E_BIN_DIR):
	@echo "Creating E2E binary directory: $(E2E_BIN_DIR)"
	@mkdir -p $(E2E_BIN_DIR)

.PHONY: build-e2e-timeserver-docker
build-e2e-timeserver-docker: tests/integration/examples/Dockerfile.timeserver tests/integration/examples/timeserver_patch/main.py
ifneq ($(SKIP_DOCKER_TESTS),true)
ifdef HAS_DOCKER
	@echo "Building E2E time server Docker image (mcpany-e2e-time-server)..."
	@$(DOCKER_BUILD_CMD) -t mcpany-e2e-time-server -f tests/integration/examples/Dockerfile.timeserver $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) tests/integration/examples
else
	@echo "Docker not found. Cannot build E2E time server image."
	@exit 1
endif
else
	@echo "Skipping E2E time server Docker image build..."
endif


# Target to build the server Docker image.
# Set PUSH=true to push the image to the registry.
# Set PLATFORMS to specify target platforms, e.g., linux/amd64,linux/arm64.
# If PLATFORMS is not set, it defaults to the host architecture.
# For single-platform local builds, --load is added automatically.
build-docker: docker/Dockerfile.server prepare-docker-context
ifdef HAS_DOCKER
	@{ \
		PLATFORMS="$(PLATFORMS)"; \
		if [ -z "$$PLATFORMS" ]; then \
			HOST_ARCH=$$(uname -m); \
			case "$$HOST_ARCH" in \
				x86_64) PLATFORMS=linux/amd64 ;; \
				aarch64) PLATFORMS=linux/arm64 ;; \
				*) PLATFORMS=linux/amd64 ;; \
			esac; \
		else \
			PLATFORMS=$$(echo "$$PLATFORMS" | tr ' ' ','); \
		fi; \
		echo "Building server Docker image with tags ($(SERVER_IMAGE_TAGS)) for platforms: $${PLATFORMS}"; \
		CACHE_ARGS=""; \
		if [ -n "$(CACHE_TO)" ]; then \
			CACHE_ARGS="--cache-to=$(CACHE_TO)"; \
		fi; \
		if [ -n "$(CACHE_FROM)" ]; then \
			CACHE_ARGS="$${CACHE_ARGS} --cache-from=$(CACHE_FROM)"; \
		fi; \
		TAG_ARGS=$$(echo "$(SERVER_IMAGE_TAGS)" | sed -e 's/ / -t /g' -e 's/^/-t /'); \
		if [ "$(PUSH)" = "true" ]; then \
			echo "Building and pushing image..."; \
			CMD_ARGS="--platform $${PLATFORMS} $${TAG_ARGS} -f docker/Dockerfile.server"; \
			$(DOCKER_BUILDX_CMD) build $${CMD_ARGS} $${CACHE_ARGS} --push .; \
		else \
			if [ "$(HAS_BUILDX)" = "1" ]; then \
				CMD_ARGS="--platform $${PLATFORMS} $${TAG_ARGS} -f docker/Dockerfile.server"; \
				if echo "$${PLATFORMS}" | grep -q ','; then \
					echo "Building multi-platform image locally (will not be loaded to docker images)..."; \
					$(DOCKER_BUILDX_CMD) build $${CMD_ARGS} $${CACHE_ARGS} .; \
				else \
					echo "Building single-platform image locally and loading to docker images..."; \
					$(DOCKER_BUILD_CMD) $${CMD_ARGS} $${CACHE_ARGS} .; \
				fi; \
			else \
				echo "Building image using legacy docker build..."; \
				$(DOCKER_BUILD_CMD) $${TAG_ARGS} -f docker/Dockerfile.server .; \
			fi; \
		fi; \
	}
	@echo "Server Docker image build complete."

else
	@echo "Docker not found. Cannot build server image."
	@exit 1
endif

# Target to build http-echo-server Docker image
build-http-echo-docker: docker/Dockerfile.http-echo-server prepare-docker-context
ifneq ($(SKIP_DOCKER_TESTS),true)
ifdef HAS_DOCKER
	@echo "Building http-echo-server Docker image (mcpany/http-echo-server:latest)..."
	@$(DOCKER_BUILD_CMD) -t mcpany/http-echo-server:latest -f docker/Dockerfile.http-echo-server $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) .

else
	@echo "Docker not found. Cannot build http-echo-server image."
	@exit 1
endif
else
	@echo "Skipping http-echo-server Docker image build..."
endif

# Target to build cowsay Docker image for integration tests
build-cowsay-docker: tests/integration/cmd/mocks/python_cowsay_server/Dockerfile
ifneq ($(SKIP_DOCKER_TESTS),true)
ifdef HAS_DOCKER
	@echo "Building cowsay Docker image (mcpany/e2e-cowsay-server:latest)..."
	@$(DOCKER_BUILD_CMD) -t mcpany/e2e-cowsay-server:latest -f tests/integration/cmd/mocks/python_cowsay_server/Dockerfile $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) tests/integration/cmd/mocks/python_cowsay_server
else
	@echo "Docker not found. Cannot build cowsay image."
	@exit 1
endif
else
	@echo "Skipping cowsay Docker image build..."
endif

# Target to build operator Docker image
docker-build-operator: prepare-docker-context
ifdef HAS_DOCKER
	@echo "Building operator Docker image (mcpany/operator:latest)..."
	@$(DOCKER_BUILD_CMD) -t mcpany/operator:latest -f ../k8s/operator/Dockerfile $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) ..
else
	@echo "Docker not found. Cannot build operator image."
	@exit 1
endif

# Aliases for consistent naming
docker-build-server: build-docker
docker-build-http-echo: build-http-echo-docker

# Target to build dev Docker image
docker-build-dev: docker/Dockerfile.dev
ifdef HAS_DOCKER
	@echo "Building dev Docker image (mcpany/dev:latest)..."
	@$(DOCKER_BUILD_CMD) -t mcpany/dev:latest -f docker/Dockerfile.dev $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) ..
else
	@echo "Docker not found. Cannot build dev image."
	@exit 1
endif

# Target to build all test images
build-test-images:
ifndef SKIP_DOCKER_REBUILD
	@$(MAKE) build-docker build-http-echo-docker build-e2e-timeserver-docker build-cowsay-docker build-everything-docker
else
	@echo "Skipping Docker image build (SKIP_DOCKER_REBUILD is set)"
endif


# Target to run the server Docker container
run-docker:
ifdef HAS_DOCKER
	@echo "Running server Docker container from image ($(SERVER_IMAGE_TAG))..."
	@echo "Exposing ports 50050 and 50051."
	@$(DOCKER_CMD) run --rm -p 50050:50050 -p 50051:50051 $(SERVER_IMAGE_TAG)
else
	@echo "Docker not found. Cannot run server container."
	@exit 1
endif

# Target to run the UI Docker smoke test
# Target to build UI Docker image (wrapper for CI)
docker-build-ui:
	$(MAKE) -C ../ui docker-build-ui

docker-smoke-test-ui:
ifdef HAS_DOCKER
	@echo "Building UI Docker image..."
	@$(DOCKER_BUILD_CMD) -t mcpany-ui -f ../ui/Dockerfile ../ui
	@echo "Running UI smoke test..."
	@chmod +x tests/smoke_test_ui.sh
	@./tests/smoke_test_ui.sh mcpany-ui 3333
else
	@echo "Docker not found. Cannot run UI smoke test."
	@exit 1
endif




build-everything-docker: docker/Dockerfile.everything
ifneq ($(SKIP_DOCKER_TESTS),true)
ifdef HAS_DOCKER
	@echo "Building everything server Docker image ($(EVERYTHING_IMAGE_TAG))..."
	@$(DOCKER_BUILD_CMD) -t $(EVERYTHING_IMAGE_TAG) -f docker/Dockerfile.everything $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) .
	@echo "Everything server Docker image build complete."
else
	@echo "Docker not found. Cannot build everything server image."
	@exit 1
endif
else
	@echo "Skipping everything server Docker image build..."
endif

BUILD_CONTEXT_DIR := .build-context

prepare-docker-context:
	@echo "Preparing build context..."
	@rm -rf $(BUILD_CONTEXT_DIR)
	@mkdir -p $(BUILD_CONTEXT_DIR)/core
	@cp ../go.mod ../go.sum $(BUILD_CONTEXT_DIR)/core/
	@cp -r ../proto $(BUILD_CONTEXT_DIR)/core/

clean-docker-context:
	@rm -rf $(BUILD_CONTEXT_DIR)

help:
	@echo "Available commands:"
	@echo "  make prepare        - Install development tools"
	@echo "  make build          - Build server binary"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linters"
	@echo "  make gen            - Generate protobuf files"
	@echo "  make build-docker   - Build server docker image"
	@echo "  make run            - Run server locally"
