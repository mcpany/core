
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

# Default test target# Run the tests in a docker container
test: build-test-docker
	# Override command to include args if provided
	# We can't easily pass args via `up`.
	# We should use `run` if we want to pass args, but `up` is good for dependencies.
	# Alternative: Set PLAYWRIGHT_ARGS env var?
	# Or use `run` but ensure dependencies are up.
	# Let's try to use PLAYWRIGHT_ARGS env var if we can't change command easily.
	# But docker-compose.test-ui.yml command is `npx playwright test`.
	# If we change it to use sh -c "npx playwright test $PLAYWRIGHT_ARGS", it might work.
	# BUT I prefer to use `run` for targeted tests.
	if [ -z '$(ARGS)' ]; then \
		docker compose -f ../docker/docker-compose.test-ui.yml up --abort-on-container-exit --exit-code-from ui-tests; \
	else \
		docker compose -f ../docker/docker-compose.test-ui.yml run --rm --service-ports ui-tests npx playwright test $(ARGS); \
	fi

# Run audit tests (generate screenshots)
audit: build-test-docker
	CAPTURE_SCREENSHOTS=true docker compose -f ../docker/docker-compose.test-ui.yml up --abort-on-container-exit --exit-code-from ui-tests
	docker compose -f ../docker/docker-compose.test-ui.yml down -v

test-ci:
	docker compose -f ../docker/docker-compose.test-ui.yml build
	docker compose -f ../docker/docker-compose.test-ui.yml up --quiet-pull --abort-on-container-exit --exit-code-from ui-tests
	docker compose -f ../docker/docker-compose.test-ui.yml down -v

# Local test target (runs natively)
test-local:
	@echo "Running UI tests locally..."
	npm install
	BACKEND_URL=http://localhost:50050 NEXT_PUBLIC_API_URL=http://localhost:50050 npx playwright test $(PLAYWRIGHT_ARGS)

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
