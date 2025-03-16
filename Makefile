.PHONY: build test test-coverage lint clean help all

GO = go
GOCOVER = $(GO) tool cover
GOTEST = $(GO) test
GOLINT = golangci-lint
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html
GO_PACKAGES = ./pkg/... ./examples/...

all: lint test build

build:
	@echo "Building Go code..."
	@$(GO) build ./...

test:
	@echo "Running tests..."
	@$(GOTEST) -v $(GO_PACKAGES)

test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(GO_PACKAGES)
	@$(GOCOVER) -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated at $(COVERAGE_HTML)"
	@$(GOCOVER) -func=$(COVERAGE_FILE)

lint:
	@echo "Running linter..."
	@go vet ./pkg/... ./examples/...
	@go fmt ./pkg/... ./examples/...

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

help:
	@echo "Available targets:"
	@echo "  all            - Run lint, test, and build"
	@echo "  build          - Build Go code"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"

.PHONY: release-patch release-minor release-major

VERSION_FILE = VERSION
CURRENT_VERSION = $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.0.0")
MAJOR = $(shell echo $(CURRENT_VERSION) | cut -d. -f1)
MINOR = $(shell echo $(CURRENT_VERSION) | cut -d. -f2)
PATCH = $(shell echo $(CURRENT_VERSION) | cut -d. -f3)

release-patch:
	@echo "Releasing patch version..."
	@NEW_PATCH=$$(( $(PATCH) + 1 )); \
	NEW_VERSION="$(MAJOR).$(MINOR).$$NEW_PATCH"; \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	git add $(VERSION_FILE); \
	git commit -m "Bump version to $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Version $$NEW_VERSION"; \
	echo "Version bumped to $$NEW_VERSION"

release-minor:
	@echo "Releasing minor version..."
	@NEW_MINOR=$$(( $(MINOR) + 1 )); \
	NEW_VERSION="$(MAJOR).$$NEW_MINOR.0"; \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	git add $(VERSION_FILE); \
	git commit -m "Bump version to $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Version $$NEW_VERSION"; \
	echo "Version bumped to $$NEW_VERSION"

release-major:
	@echo "Releasing major version..."
	@NEW_MAJOR=$$(( $(MAJOR) + 1 )); \
	NEW_VERSION="$$NEW_MAJOR.0.0"; \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	git add $(VERSION_FILE); \
	git commit -m "Bump version to $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Version $$NEW_VERSION"; \
	echo "Version bumped to $$NEW_VERSION" 