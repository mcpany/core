
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

.PHONY: test test-local build-test-docker audit

TIMESTAMP := $(shell date +%s)
SERVER_IMAGE_TAG := ui-test-$(TIMESTAMP)
SERVER_IMAGE ?= mcpany/server:$(SERVER_IMAGE_TAG)

TEST_UID := $(shell id -u)
TEST_GID := $(shell id -g)
export TEST_UID
export TEST_GID

# Default test target
# Run the tests in a docker container
test: build-test-docker build-server-docker
	mkdir -p test-results .audit ../build/playwright-report
	chmod -R 777 test-results .audit ../build/playwright-report || true
	if [ -z '$(ARGS)' ]; then \
		SERVER_IMAGE=$(SERVER_IMAGE) docker compose -f ./docker-compose.test.yml up --abort-on-container-exit --exit-code-from ui-tests; \
	else \
		SERVER_IMAGE=$(SERVER_IMAGE) docker compose -f ./docker-compose.test.yml run --rm --service-ports ui-tests npx playwright test $(ARGS); \
	fi

build-server-docker:
	@echo "Building Server Docker image ($(SERVER_IMAGE)) via server/Makefile..."
	$(MAKE) -C ../server build-docker
	docker tag mcpany/server:latest $(SERVER_IMAGE)

# Run audit tests (generate screenshots)
audit: build-test-docker
	SERVER_IMAGE=$(SERVER_IMAGE) CAPTURE_SCREENSHOTS=true docker compose -f ./docker-compose.test.yml up --abort-on-container-exit --exit-code-from ui-tests
	docker compose -f ./docker-compose.test.yml down -v

test-ci:
	mkdir -p test-results .audit ../build/playwright-report
	chmod -R 777 test-results .audit ../build/playwright-report || true
	docker compose -f ./docker-compose.test.yml up --quiet-pull --abort-on-container-exit --exit-code-from ui-tests
	docker compose -f ./docker-compose.test.yml down -v

# Local test target (runs natively)
test-local:
	@echo "Running UI tests locally..."
	npm install
	BACKEND_URL=http://localhost:50050 NEXT_PUBLIC_API_URL=http://localhost:50050 npx playwright test $(PLAYWRIGHT_ARGS)

# Build the UI test Docker image
build-test-docker:
	@echo "Building UI test Docker image (mcpany-ui-e2e)..."
	@$(DOCKER_BUILD_CMD) -t mcpany-ui-e2e -f Dockerfile.e2e $(if $(CACHE_TO),--cache-to=$(CACHE_TO)) $(if $(CACHE_FROM),--cache-from=$(CACHE_FROM)) .

docker-build-ui:
	@echo "Building UI Docker image (mcpany/ui:latest)..."
	@$(DOCKER_BUILD_CMD) -t mcpany/ui:latest -f Dockerfile .

# Build the UI production build
build:
	@echo "Building UI..."
	npm install && npm run build

update-screenshots:
	@echo "Updating screenshots..."
	docker compose -f ./docker-compose.test.yml run --rm ui-tests npx playwright test --config playwright.screenshots.config.ts
