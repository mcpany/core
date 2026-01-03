
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

# Default test target (runs tests in Docker Compose)
test:
	docker compose -f ../docker/docker-compose.test-ui.yml up --build --abort-on-container-exit --exit-code-from ui-tests
	docker compose -f ../docker/docker-compose.test-ui.yml down -v

test-ci:
	docker compose -f ../docker/docker-compose.test-ui.yml up --abort-on-container-exit --exit-code-from ui-tests
	docker compose -f ../docker/docker-compose.test-ui.yml down -v

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

update-screenshots:
	@echo "Updating screenshots..."
	docker compose -f ../docker/docker-compose.test-ui.yml run --rm ui-tests npx playwright test --config playwright.screenshots.config.ts
