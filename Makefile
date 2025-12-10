.PHONY: help build build-api build-grading build-worker build-database run-api run-grading run-worker clean test fmt vet docker-build docker-up docker-down docker-logs docker-restart deps tidy

# Variables
BINARY_DIR=bin
API_BINARY=$(BINARY_DIR)/api
GRADING_BINARY=$(BINARY_DIR)/grading
WORKER_BINARY=$(BINARY_DIR)/worker
DATABASE_BINARY=$(BINARY_DIR)/database

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod

# Colors for terminal output
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

## help: Display this help message
help:
	@echo "$(GREEN)Mithril - Available Make Targets:$(NC)"
	@echo ""
	@echo "$(YELLOW)Building:$(NC)"
	@echo "  make build              - Build all services"
	@echo "  make build-api          - Build API service"
	@echo "  make build-grading      - Build grading service"
	@echo "  make build-worker       - Build worker service"
	@echo "  make build-database     - Build database migrations"
	@echo ""
	@echo "$(YELLOW)Running (local):$(NC)"
	@echo "  make run-api            - Run API service locally"
	@echo "  make run-grading        - Run grading service locally"
	@echo "  make run-worker         - Run worker service locally"
	@echo ""
	@echo "$(YELLOW)Testing & Quality:$(NC)"
	@echo "  make test               - Run tests"
	@echo "  make fmt                - Format code"
	@echo "  make vet                - Run go vet"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  make docker-build       - Build all Docker images"
	@echo "  make docker-up          - Start all services with Docker Compose"
	@echo "  make docker-down        - Stop all services"
	@echo "  make docker-logs        - Show logs from all services"
	@echo "  make docker-restart     - Restart all services"
	@echo ""
	@echo "$(YELLOW)Dependencies:$(NC)"
	@echo "  make deps               - Download dependencies"
	@echo "  make tidy               - Tidy go.mod"
	@echo ""
	@echo "$(YELLOW)Cleanup:$(NC)"
	@echo "  make clean              - Remove binaries and clean build cache"

## build: Build all services
build: build-api build-grading build-worker build-database
	@echo "$(GREEN)All services built successfully$(NC)"

## build-api: Build API service
build-api:
	@echo "$(YELLOW)Building API service...$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(API_BINARY) ./cmd/api
	@echo "$(GREEN)API service built: $(API_BINARY)$(NC)"

## build-grading: Build grading service
build-grading:
	@echo "$(YELLOW)Building grading service...$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(GRADING_BINARY) ./cmd/grading
	@echo "$(GREEN)Grading service built: $(GRADING_BINARY)$(NC)"

## build-worker: Build worker service
build-worker:
	@echo "$(YELLOW)Building worker service...$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(WORKER_BINARY) ./cmd/worker
	@echo "$(GREEN)Worker service built: $(WORKER_BINARY)$(NC)"

## build-database: Build database migrations
build-database:
	@echo "$(YELLOW)Building database migrations...$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(DATABASE_BINARY) ./cmd/database
	@echo "$(GREEN)Database migrations built: $(DATABASE_BINARY)$(NC)"

## run-api: Run API service locally
run-api: build-api
	@echo "$(YELLOW)Running API service...$(NC)"
	./$(API_BINARY)

## run-grading: Run grading service locally
run-grading: build-grading
	@echo "$(YELLOW)Running grading service...$(NC)"
	./$(GRADING_BINARY)

## run-worker: Run worker service locally
run-worker: build-worker
	@echo "$(YELLOW)Running worker service...$(NC)"
	./$(WORKER_BINARY)

## test: Run tests
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GOTEST) -v ./...

## fmt: Format code
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GOVET) ./...

## docker-build: Build all Docker images
docker-build:
	@echo "$(YELLOW)Building Docker images...$(NC)"
	docker-compose build

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "$(YELLOW)Starting services with Docker Compose...$(NC)"
	docker compose up -d
	@echo "$(GREEN)Services started. Use 'make docker-logs' to view logs$(NC)"

## docker-down: Stop all services
docker-down:
	@echo "$(YELLOW)Stopping services...$(NC)"
	docker-compose down

## docker-logs: Show logs from all services
docker-logs:
	docker-compose logs -f

## docker-restart: Restart all services
docker-restart: docker-down docker-up
	@echo "$(GREEN)Services restarted$(NC)"

## deps: Download dependencies
deps:
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	$(GOGET) -v ./...

## tidy: Tidy go.mod
tidy:
	@echo "$(YELLOW)Tidying go.mod...$(NC)"
	$(GOMOD) tidy

## clean: Remove binaries and clean build cache
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	@echo "$(GREEN)Clean complete$(NC)"
