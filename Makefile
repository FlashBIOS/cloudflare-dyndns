# Go commands and settings. Adjust these variables as needed.
GOCMD      = go
GOBUILD    = $(GOCMD) build
GOTEST     = $(GOCMD) test
GOFMT      = $(GOCMD) fmt
GOVET      = $(GOCMD) vet
GOMOD      = $(GOCMD) mod

# The name of the output binary. Adjust if your main package uses a different name.
BINARY_NAME = cloudflare-dyndns

# The directory where the binary will be placed.
BUILD_DIR = bin

# Set default target OS's if none are specified.
TARGETS ?= darwin linux windows

# Set the build architecture.
ARCH ?= amd64 arm64

.PHONY: all release build test run fmt vet clean tidy verify

# Default target builds the binary.
all: build

# Release compiles an optimized version of the binary.
release: tidy verify test clean
	@echo "Building binaries for: $(TARGETS)"
	@for os in $(TARGETS); do \
		for arch in $(ARCH); do \
			echo "Building for $$os $$arch..."; \
			mkdir -p $(BUILD_DIR)/$$os; \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$$os/$$arch/$(BINARY_NAME) .; \
		done; \
	done

# Build compiles the Go code and outputs the binary into the build directory.
build:
	@echo "Building binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)

# Test runs all tests.
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run builds and runs the binary.
run: build
	@echo "Running binary..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Fmt formats the Go code.
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Vet reports any suspicious constructs in the code.
vet:
	@echo "Linting with vet..."
	$(GOVET) ./...

# Clean removes build artifacts.
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Tidy cleans up the mod file.
tidy:
	@echo "Tidying up the go.mod file..."
	$(GOMOD) tidy

# Verify all the module dependencies
verify:
	@echo "Verifying the module dependencies..."
	$(GOMOD) verify