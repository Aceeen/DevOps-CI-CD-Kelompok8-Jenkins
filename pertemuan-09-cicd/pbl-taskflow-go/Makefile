# Makefile - TaskFlow API
BINARY          = bin/taskflow-api
IMAGE           = taskflow-api
REGISTRY       ?= docker.io/your-dockerhub-username
VERSION        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DB_URL         ?= postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable
APP_PORT       ?= 18080
APP_BASE_URL   ?= http://localhost:$(APP_PORT)
CONTAINER_NAME ?= taskflow-api-staging

.PHONY: all vet test test-race test-cover test-integration build \
        docker-build docker-push docker-stable docker-run smoke-test \
        docker-size-report rollback db-up db-down up clean help

all: vet test build

## go vet - static analysis
vet:
	@echo "-> go vet ./..."
	go vet ./...

## Run unit tests
test:
	@echo "-> go test ./..."
	go test ./... -v -timeout 30s

## Run tests with race detector
test-race:
	@echo "-> go test -race ./..."
	go test ./... -race -timeout 30s

## Coverage report
test-cover:
	@echo "-> coverage report"
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

## Integration test (requires DB_URL / running postgres)
test-integration:
	@echo "-> integration test (DATABASE_URL=$(DB_URL))"
	DATABASE_URL=$(DB_URL) go test -tags=integration ./... -v -race -timeout 60s

## Build Linux binary
build:
	@echo "-> go build ($(VERSION))"
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build \
		-ldflags="-w -s" \
		-o $(BINARY) ./cmd/server

## Build Docker image with SHA tag
docker-build:
	@echo "-> docker build ($(VERSION))"
	docker build -t $(REGISTRY)/$(IMAGE):sha-$(VERSION) .
	@docker images $(REGISTRY)/$(IMAGE):sha-$(VERSION) --format "Size: {{.Size}}"

## Push SHA image to registry
docker-push:
	@echo "-> docker push $(REGISTRY)/$(IMAGE):sha-$(VERSION)"
	docker push $(REGISTRY)/$(IMAGE):sha-$(VERSION)

## Tag current SHA image as stable
docker-stable:
	@echo "-> tag $(VERSION) as stable"
	docker tag $(REGISTRY)/$(IMAGE):sha-$(VERSION) $(REGISTRY)/$(IMAGE):stable
	docker push $(REGISTRY)/$(IMAGE):stable

## Run the pushed image locally for staging smoke test
docker-run:
	@echo "-> run $(REGISTRY)/$(IMAGE):sha-$(VERSION) on port $(APP_PORT)"
	docker rm -f $(CONTAINER_NAME) 2>/dev/null || true
	docker run -d --rm \
	  --name $(CONTAINER_NAME) \
	  -p $(APP_PORT):8080 \
	  $(REGISTRY)/$(IMAGE):sha-$(VERSION)
	docker ps --filter name=$(CONTAINER_NAME)

## Smoke test staging endpoints
smoke-test:
	@echo "-> smoke test $(APP_BASE_URL)"
	@sleep 5
	curl -f $(APP_BASE_URL)/health
	curl -f $(APP_BASE_URL)/api/v1/stats
	@echo "Smoke test passed."

## Compare multi-stage and legacy image sizes
docker-size-report:
	@echo "-> build multi-stage and legacy images for comparison"
	docker build -t $(REGISTRY)/$(IMAGE):sha-$(VERSION) .
	docker build -f Dockerfile.legacy -t $(IMAGE):legacy-compare .
	@multi_human=$$(docker image ls $(REGISTRY)/$(IMAGE):sha-$(VERSION) --format '{{.Size}}'); \
	multi_bytes=$$(docker image inspect $(REGISTRY)/$(IMAGE):sha-$(VERSION) --format '{{.Size}}'); \
	legacy_human=$$(docker image ls $(IMAGE):legacy-compare --format '{{.Size}}'); \
	legacy_bytes=$$(docker image inspect $(IMAGE):legacy-compare --format '{{.Size}}'); \
	delta_bytes=$$((legacy_bytes - multi_bytes)); \
	printf "%s\n" \
	  "Multi-stage image: $(REGISTRY)/$(IMAGE):sha-$(VERSION)" \
	  "  size_human: $$multi_human" \
	  "  size_bytes: $$multi_bytes" \
	  "Legacy image: $(IMAGE):legacy-compare" \
	  "  size_human: $$legacy_human" \
	  "  size_bytes: $$legacy_bytes" \
	  "Difference_bytes: $$delta_bytes" \
	  > image-size-report.txt
	@cat image-size-report.txt

## Rollback: run a previously published image tag
## Usage: make rollback ROLLBACK_TAG=sha-a3f2c1d
rollback:
	@test -n "$(ROLLBACK_TAG)" || (echo "ERROR: Set ROLLBACK_TAG=sha-xxxxx"; exit 1)
	@echo "-> rolling back to $(REGISTRY)/$(IMAGE):$(ROLLBACK_TAG)"
	docker pull $(REGISTRY)/$(IMAGE):$(ROLLBACK_TAG)
	docker stop taskflow-api 2>/dev/null || true
	docker run -d --rm \
	  --name taskflow-api \
	  -p 8080:8080 \
	  -e DATABASE_URL=$(DB_URL) \
	  $(REGISTRY)/$(IMAGE):$(ROLLBACK_TAG)
	@echo "-> waiting for server..."
	@sleep 5
	curl -sf http://localhost:8080/health || (echo "ERROR: health check failed"; exit 1)
	@echo "Rollback successful to $(ROLLBACK_TAG)"

## Start postgres only
db-up:
	docker compose up -d postgres
	@echo "-> waiting for postgres..."
	@sleep 3
	@echo "Postgres is ready on localhost:5432"

## Start full stack
up:
	docker compose up -d
	@echo "Stack is running. API: http://localhost:8080/health"

## Stop stack
db-down:
	docker compose down

## Clean local artifacts
clean:
	rm -rf bin/ coverage.out image-size-report.txt

## Show available targets
help:
	@grep -E '^##' Makefile | sed 's/## /  /'
