
# UI Makefile

# Variables
DOCKER_CMD := docker
DOCKER_BUILDX_CMD := docker buildx
# Check if docker needs sudo (simple check)
ifeq ($(shell docker info >/dev/null 2>&1 && echo 1 || echo 0), 0)
	DOCKER_CMD := sudo docker
	DOCKER_BUILDX_CMD := sudo docker buildx
endif

DOCKER_BUILD_CMD := $(DOCKER_BUILDX_CMD) build --load
# Check if buildx exists, otherwise fallback to standard build
ifeq ($(shell $(DOCKER_CMD) buildx version >/dev/null 2>&1 && echo 1 || echo 0), 0)
	DOCKER_BUILD_CMD := $(DOCKER_CMD) build
endif

.PHONY: test test-local build-test-docker

# Default test target (runs in Docker)
test: build-test-docker
	@echo "Running UI tests (Playwright) via Docker image (mcpany-ui-e2e)..."
	@if [ -n "$(PLAYWRIGHT_ARGS)" ]; then echo "Running with args: $(PLAYWRIGHT_ARGS)"; fi
	@$(DOCKER_CMD) run --rm --ipc=host --network=host \
		-v $(abspath .):/work/ui \
		-v $(abspath ../proto):/work/proto \
		-w /work/ui \
		-e CI=$(CI) \
		mcpany-ui-e2e \
		/bin/bash -c "npm install && npx playwright test $(PLAYWRIGHT_ARGS)"

# Local test target (runs natively)
test-local:
	@echo "Running UI tests locally..."
	npm install && npx playwright test $(PLAYWRIGHT_ARGS)

# Build the UI test Docker image
build-test-docker:
	@echo "Building UI test Docker image (mcpany-ui-e2e)..."
	@$(DOCKER_BUILD_CMD) -t mcpany-ui-e2e -f Dockerfile.e2e $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) .

# Build the UI production build
build:
	@echo "Building UI..."
	npm install && npm run build
