.PHONY: all build test clean run init migrate seed

# Default target
all: build

# Build the application
build:
	go build -o backend_go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f backend_go
	rm -f *.exe
	rm -f *.test
	rm -f *.out

# Run the application
run:
	go run main.go

# Initialize database
init:
	go run magefiles/main.go initdb

# Run migrations
migrate:
	go run magefiles/main.go migrate

# Seed database
seed:
	go run magefiles/main.go seed

# Reset database (clean start)
reset: init migrate seed

# Install dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Generate swagger docs (if you add swagger generation in future)
swagger:
	@echo "Swagger generation not implemented yet"

# Help target
help:
	@echo "Available targets:"
	@echo "  all          : Build the application (default)"
	@echo "  build        : Build the application"
	@echo "  test         : Run tests"
	@echo "  test-coverage: Run tests with coverage"
	@echo "  clean        : Clean build artifacts"
	@echo "  run          : Run the application"
	@echo "  init         : Initialize database"
	@echo "  migrate      : Run migrations"
	@echo "  seed         : Seed database"
	@echo "  reset        : Reset database (clean start)"
	@echo "  deps         : Install dependencies"
	@echo "  fmt          : Format code"
	@echo "  lint         : Run linter"
	@echo "  swagger      : Generate swagger docs (not implemented)"
	@echo "  help         : Show this help message" 