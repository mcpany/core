# Makefile

# Variables
GO = go
GO_ENV := GOCACHE=/tmp/.gocache GOMODCACHE=/tmp/.modcache
GO_CMD := $(GO_ENV) $(GO)
SERVER_IMAGE_TAG ?= mcpxy/server:latest

HAS_DOCKER := $(shell command -v docker 2> /dev/null)
# Check if docker can be run without sudo
ifeq ($(shell docker info >/dev/null 2>&1; echo $$?), 0)
	DOCKER_CMD := docker
	DOCKER_BUILDX_CMD := docker buildx
	SUDO_MSG := " (no sudo)"
	NEEDS_SUDO_FOR_DOCKER := 0
else
	DOCKER_CMD := sudo docker
	DOCKER_BUILDX_CMD := sudo docker buildx
	NEEDS_SUDO_FOR_DOCKER := 1
endif

# Variables for protoc installation
PROTOC_VERSION_URL := https://api.github.com/repos/protocolbuffers/protobuf/releases/latest
PROTOC_DOWNLOAD_URL_BASE := https://github.com/protocolbuffers/protobuf/releases/download
PROTOC_GEN_GO_VERSION ?= latest
PROTOC_GEN_GO_GRPC_VERSION ?= latest
PROTOC_ZIP := protoc.zip
PROTOC_INSTALL_DIR := $(CURDIR)/build/env/bin
PROTOC_VERSION := v32.1
LOCAL_BIN_DIR := $(CURDIR)/build/bin
PROTOC_GEN_GO := $(PROTOC_INSTALL_DIR)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(PROTOC_INSTALL_DIR)/protoc-gen-go-grpc
PROTOC_BIN := $(PROTOC_INSTALL_DIR)/protoc
GOLANGCI_LINT_BIN := $(PROTOC_INSTALL_DIR)/golangci-lint
GOFUMPT_BIN := $(PROTOC_INSTALL_DIR)/gofumpt
GOIMPORTS_BIN := $(PROTOC_INSTALL_DIR)/goimports
PRE_COMMIT_VERSION := 4.3.0
PRE_COMMIT_BIN := $(PROTOC_INSTALL_DIR)/pre-commit
# ==============================================================================
# Release Targets
# ==============================================================================
RELEASE_DIR := $(CURDIR)/build/release
# PLATFORMS variable will be used to define the target platforms for the build.
# Example: PLATFORMS := linux/amd64 linux/arm64
PLATFORMS ?= linux/amd64 linux/386 linux/arm64 linux/arm

# Find all .proto files, excluding vendor/cache directories
PROTO_FILES := $(shell find proto -name "*.proto")

.PHONY: all gen build test e2e clean run build-docker run-docker gen build test e2e-local check-local release release-local release-docker

all: build

# ==============================================================================
# Main Targets (Default to local, use USE_DOCKER=1 to switch to Docker)
# ==============================================================================

release:
ifeq ($(USE_DOCKER), 1)
	@$(MAKE) release-docker
else
	@$(MAKE) release-local
endif

# ==============================================================================
# Local Commands
# ==============================================================================

prepare:
	@echo "Preparing development environment..."
	@mkdir -p $(LOCAL_BIN_DIR)
	@mkdir -p $(PROTOC_INSTALL_DIR)
	@# Check if protoc is installed
	@export PATH=$(PROTOC_INSTALL_DIR):$$PATH; \
	PROTOC_TAG=$(PROTOC_VERSION); \
	if [ -f "$(PROTOC_BIN)" ]; then \
		INSTALLED_VERSION=v$$($(PROTOC_BIN) --version | sed 's/libprotoc //'); \
		if [ "$${INSTALLED_VERSION}" = "$${PROTOC_TAG}" ]; then \
			echo "protoc version $${INSTALLED_VERSION} is already installed."; \
		else \
			echo "protoc version mismatch. Installed: $${INSTALLED_VERSION}, Required: $${PROTOC_TAG}. Re-installing..."; \
			rm -f "$(PROTOC_BIN)"; \
			$(MAKE) prepare; \
		fi; \
	else \
		echo "protoc not found, attempting to install version $${PROTOC_TAG}..."; \
		if ! command -v curl >/dev/null 2>&1 || ! command -v unzip >/dev/null 2>&1; then \
			echo "Error: curl and unzip are required to download protoc. Please install them and try again."; \
			exit 1; \
		fi; \
		PROTOC_VERSION_NO_V=$$(echo "$${PROTOC_TAG}" | sed 's/v//'); \
		PROTOC_DOWNLOAD_URL_NO_V="$(PROTOC_DOWNLOAD_URL_BASE)/$${PROTOC_TAG}/protoc-$${PROTOC_VERSION_NO_V}-linux-x86_64.zip"; \
		echo "Downloading protoc from $${PROTOC_DOWNLOAD_URL_NO_V}..."; \
		if curl -sSL "$${PROTOC_DOWNLOAD_URL_NO_V}" -o "$(PROTOC_ZIP)"; then \
			echo "Unzipping to $(PROTOC_INSTALL_DIR)..."; \
			unzip -o "$(PROTOC_ZIP)" -d "$(PROTOC_INSTALL_DIR)"; \
			mv "$(PROTOC_INSTALL_DIR)/bin/protoc" "$(PROTOC_BIN)"; \
			if [ -f "$(PROTOC_BIN)" ]; then \
				export PATH=$(PROTOC_INSTALL_DIR):$$PATH; \
				echo "protoc installed successfully to $(PROTOC_INSTALL_DIR). This directory has been added to PATH for this session."; \
				echo "You may want to add it to your system PATH permanently: export PATH=$(PROTOC_INSTALL_DIR):$$PATH"; \
				$(PROTOC_BIN) --version; \
			else \
				echo "Error: protoc binary not found in $(PROTOC_INSTALL_DIR) after unzip. The downloaded archive may not have the expected structure."; \
				exit 1; \
			fi; \
			rm -f "$(PROTOC_ZIP)"; \
		fi; \
	fi
	@# Install Go protobuf plugins
	@echo "Installing Go protobuf plugins..."
	@GOBIN=$(PROTOC_INSTALL_DIR) $(GO_CMD) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@GOBIN=$(PROTOC_INSTALL_DIR) $(GO_CMD) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@# Install golangci-lint
	@echo "Checking for golangci-lint..."
	@if [ -f "$(GOLANGCI_LINT_BIN)" ]; then \
		echo "golangci-lint is already installed."; \
	else \
		echo "Installing golangci-lint..."; \
		GOBIN=$(PROTOC_INSTALL_DIR) $(GO_CMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if [ ! -f "$(GOLANGCI_LINT_BIN)" ]; then \
		echo "golangci-lint not found at $(GOLANGCI_LINT_BIN) after attempting install. Please check your GOPATH/GOBIN setup and PATH."; \
		exit 1; \
	fi
	@echo "Checking for Go protobuf plugins..."
	@if [ ! -f "$(PROTOC_GEN_GO)" ]; then \
		echo "protoc-gen-go not found at $(PROTOC_GEN_GO) after attempting install. Please check your GOPATH/GOBIN setup and PATH."; \
		exit 1; \
	fi
	@if [ ! -f "$(PROTOC_GEN_GO_GRPC)" ]; then \
		echo "protoc-gen-go-grpc not found at $(PROTOC_GEN_GO_GRPC) after attempting install. Please check your GOPATH/GOBIN setup and PATH."; \
		exit 1; \
	fi
	@echo "Go protobuf plugins installation check complete."
	@# Install Python dependencies and pre-commit hooks
	@echo "Checking for Python to install dependencies and pre-commit hooks..."
	@if command -v python >/dev/null 2>&1; then \
		PYTHON_CMD=python; \
	elif command -v python3 >/dev/null 2>&1; then \
		PYTHON_CMD=python3; \
	else \
		PYTHON_CMD=""; \
	fi; \
	if [ -n "$$PYTHON_CMD" ]; then \
		echo "Python found. Installing/updating pre-commit and fastmcp..."; \
		VENV_DIR=$(CURDIR)/build/venv; \
		$$PYTHON_CMD -m venv $$VENV_DIR; \
		$$VENV_DIR/bin/pip install --upgrade pip; \
		$$VENV_DIR/bin/pip install "pre-commit==$(PRE_COMMIT_VERSION)"; \
		$$VENV_DIR/bin/pip install "fastmcp>=2.0.0" --upgrade; \
		$$VENV_DIR/bin/pre-commit install || true; \
	else \
		echo "Python not found, skipping Python dependency installation and pre-commit hook setup."; \
	fi
	@echo "Preparation complete."


gen: prepare
	@echo "Removing old protobuf files..."
	@-find . -name "*.pb.go" -delete
	@echo "Generating protobuf files..."
	@export PATH=$(PROTOC_INSTALL_DIR):$$PATH; \
		echo "Using protoc: $$(protoc --version)"; \
		mkdir -p build; \
		find proto -name "*.proto" -exec protoc --experimental_editions=true \
			--proto_path=. \
			--descriptor_set_out=build/all.protoset \
			--include_imports \
			--go_out=. \
			--go_opt=module=github.com/mcpxy/core,default_api_level=API_OPAQUE \
			--go-grpc_out=. \
			--go-grpc_opt=module=github.com/mcpxy/core \
			{} +
	@echo "Protobuf generation complete."

build: gen
	@echo "Building Go project locally..."
	@$(GO_CMD) build -buildvcs=false -o ./build/bin/server ./cmd/server

test: build build-examples build-e2e-mocks build-e2e-timeserver-docker
	@echo "Installing Python dependencies for command example..."
	@python3 -m pip install -r examples/upstream/command/server/requirements.txt
	@echo "Running Go tests locally with a 300s timeout and coverage..."
	@MCPXY_DEBUG=true CGO_ENABLED=1 USE_SUDO_FOR_DOCKER=$(NEEDS_SUDO_FOR_DOCKER) $(GO_CMD) test -v -count=1 -timeout 300s -tags=e2e -cover -coverprofile=coverage.out ./...

test-fast: gen build build-examples build-e2e-mocks build-e2e-timeserver-docker
	@echo "Running fast Go tests locally with a 300s timeout..."
	@MCPXY_DEBUG=true CGO_ENABLED=1 USE_SUDO_FOR_DOCKER=$(NEEDS_SUDO_FOR_DOCKER) $(GO_CMD) test -v -count=1 -timeout 300s ./...

# ==============================================================================
# Example Binaries Build
# ==============================================================================
EXAMPLE_BIN_DIR := $(CURDIR)/build/examples/bin

# List of example binaries
.PHONY: build-examples build-calculator-stdio
build-examples: build-calculator-stdio

build-calculator-stdio:
	@echo "Building example service: calculator-stdio"
	@mkdir -p $(EXAMPLE_BIN_DIR)
	@$(GO_CMD) build -buildvcs=false -o $(EXAMPLE_BIN_DIR)/calculator-stdio ./tests/integration/calculator/cmd/stdio/main.go

# ==============================================================================
# Other Commands
# ==============================================================================

lint: gen
	@echo "Cleaning golangci-lint cache..."
	@$(GOLANGCI_LINT_BIN) cache clean
	@echo "Running golangci-lint with fix..."
	@$(GOLANGCI_LINT_BIN) run --fix ./...

clean:
	@echo "Cleaning generated protobuf files and build artifacts..."
	@-find . -name "*.pb.go" -delete
	@rm -rf build

run: build
	@echo "Starting MCP-XY server locally..."
	@./build/bin/server

# ==============================================================================
# E2E Test Related Builds
# ==============================================================================
E2E_MOCK_DIR := tests/integration/cmd/mocks
E2E_BIN_DIR := $(CURDIR)/build/test/bin

# List of mock service directories (which are also their binary names)
E2E_MOCK_SERVICES := http_echo_server http_authed_echo_server grpc_calculator_server grpc_authed_calculator_server openapi_calculator_server websocket_echo_server webrtc_echo_server

# Target to build all E2E mock services
.PHONY: build-e2e-mocks
build-e2e-mocks: $(E2E_BIN_DIR) $(addprefix $(E2E_BIN_DIR)/,$(E2E_MOCK_SERVICES))

# Rule to build a single E2E mock service
# < is the first prerequisite (the main.go file)
# * is the stem of the pattern match (the service name)
# @ is the target name (the output binary path)
$(E2E_BIN_DIR)/%: $(E2E_MOCK_DIR)/%/main.go
	@echo "Building E2E mock service: $* from $< into $(E2E_BIN_DIR)"
	@$(GO_CMD) build -buildvcs=false -o $@ $<

# Target to ensure the E2E binary directory exists
$(E2E_BIN_DIR):
	@echo "Creating E2E binary directory: $(E2E_BIN_DIR)"
	@mkdir -p $(E2E_BIN_DIR)

.PHONY: build-e2e-timeserver-docker
build-e2e-timeserver-docker: tests/integration/examples/Dockerfile.timeserver tests/integration/examples/timeserver_patch/main.py
ifdef HAS_DOCKER
	@echo "Building E2E time server Docker image (mcpxy-e2e-time-server)..."
	-@$(DOCKER_CMD) build -t mcpxy-e2e-time-server -f tests/integration/examples/Dockerfile.timeserver tests/integration/examples
else
	@echo "Docker not found. Cannot build E2E time server image."
	@exit 1
endif


# Target to build the server Docker image.
# Set PUSH=true to push the image to the registry.
# Set PLATFORMS to specify target platforms, e.g., linux/amd64,linux/arm64.
# If PLATFORMS is not set, it defaults to the host architecture.
# For single-platform local builds, --load is added automatically.
build-docker: docker/Dockerfile.server
ifdef HAS_DOCKER
	@{ \
		if [ -z "$(PLATFORMS)" ]; then \
			HOST_ARCH=$$(uname -m); \
			case "$$HOST_ARCH" in \
				x86_64) PLATFORMS=linux/amd64 ;; \
				aarch64) PLATFORMS=linux/arm64 ;; \
				*) PLATFORMS=linux/amd64 ;; \
			esac; \
		fi; \
		echo "Building server Docker image ($(SERVER_IMAGE_TAG)) for platforms: $${PLATFORMS}"; \
		CMD_ARGS="--platform $${PLATFORMS} -t $(SERVER_IMAGE_TAG) -f docker/Dockerfile.server ."; \
		if [ "$(PUSH)" = "true" ]; then \
			echo "Building and pushing image..."; \
			$(DOCKER_BUILDX_CMD) build $${CMD_ARGS} --push; \
		else \
			if echo "$${PLATFORMS}" | grep -q ','; then \
				echo "Building multi-platform image locally (will not be loaded to docker images)..."; \
				$(DOCKER_BUILDX_CMD) build $${CMD_ARGS}; \
			else \
				echo "Building single-platform image locally and loading to docker images..."; \
				$(DOCKER_BUILDX_CMD) build $${CMD_ARGS} --load; \
			fi; \
		fi; \
	}
	@echo "Server Docker image build complete."
else
	@echo "Docker not found. Cannot build server image."
	@exit 1
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

EVERYTHING_IMAGE_TAG ?= mcpxy/everything:latest
build-everything-docker: docker/Dockerfile.everything
ifdef HAS_DOCKER
	@echo "Building everything server Docker image ($(EVERYTHING_IMAGE_TAG))..."
	@$(DOCKER_BUILDX_CMD) build --load -t $(EVERYTHING_IMAGE_TAG) -f docker/Dockerfile.everything .
	@echo "Everything server Docker image build complete."
else
	@echo "Docker not found. Cannot build everything server image."
	@exit 1
endif

release-local: prepare
	@echo "Building release binaries for platforms: $(PLATFORMS)..."
	@mkdir -p $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		echo "Building for $${GOOS}/$${GOARCH}..."; \
		CGO_ENABLED=0 GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO_CMD) build -a -installsuffix cgo -buildvcs=false -o $(RELEASE_DIR)/server-$${GOOS}-$${GOARCH} ./cmd/server; \
	done
	@echo "Release binaries are in $(RELEASE_DIR)"

# The release-docker target will build and push multi-architecture Docker images.
# The GITHUB_TOKEN environment variable needs to be set for authentication.
release-docker:
	@echo "Building and pushing release images to registry..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN is not set. Please set it to your GitHub personal access token."; \
		exit 1; \
	fi
	@echo "$(GITHUB_TOKEN)" | docker login ghcr.io -u $(shell git config user.name) --password-stdin
	@$(MAKE) build-docker PUSH=true PLATFORMS=linux/amd64,linux/arm64
