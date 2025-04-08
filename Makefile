# Go commands and settings. Adjust these variables as needed.
GOCMD      = go
GOBUILD    = $(GOCMD) build
GOTEST     = $(GOCMD) test
GOFMT      = $(GOCMD) fmt
GOVET      = $(GOCMD) vet
GOMOD      = $(GOCMD) mod
GOINSTALL  = $(GOCMD) install

# The name of the output binary. Adjust if your main package uses a different name.
BINARY_NAME = cloudflare-dyndns

# The directory where the binary will be placed.
BUILD_DIR = bin
RELEASE_DIR = $(BUILD_DIR)/release

# List of GOOS/GOARCH combos (dash-separated)
GO_TARGETS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

# Set the version information.
VERSION = $(shell git tag --sort=-v:refname | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+$$" | head -n 1)
MODULE = $(shell head -1 go.mod | awk '/^module/ {print $$2; exit}')

.PHONY: all release build test run fmt vet clean tidy verify install release-all checkout-master uninstall check fmt-check

checkout-master:
	@echo "Checking out the master branch"
	git checkout master

install: checkout-master
	@echo "Installing $(BINARY_NAME) for your system"
	$(GOINSTALL) -trimpath -ldflags="-s -w -X cmd.Version=$(VERSION)" .
	@echo "Done! Don't forget to create your configuration file (see README.md) before running."

uninstall:
	rm "$(GOPATH)/bin/$(BINARY_NAME)"
	@echo "Done! Don't forget to delete any configuration file you created."

release-all: checkout-master vet verify test clean $(addprefix release-,$(GO_TARGETS))
	echo "Done!"

# Allows for parallel builds with `make -j`
release-%:
	@GOOS=$(word 1,$(subst -, ,$*)) && \
	GOARCH=$(word 2,$(subst -, ,$*)) && \
	OUTDIR=bin/release/$$GOOS/$$GOARCH && \
	echo "Building $$GOOS/$$GOARCH..." && \
	mkdir -p $$OUTDIR && \
	CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH \
	go build -ldflags "-s -w -X main.version=$(RELEASE_VERSION)" -o $$OUTDIR/cloudflare-dyndns

# Define target-specific variables for 'release'
release: os := $(shell go env GOOS)
release: arch := $(shell go env GOARCH)
release: ext := $(if $(filter windows,$(os)),.exe,)

release: checkout-master vet verify test clean
	@echo "Building binary for: $(os) $(arch)..."
	GOOS=$(os) GOARCH=$(arch) $(GOBUILD) -trimpath -ldflags="-s -w -X $(MODULE)/cmd.Version=$(VERSION)" -o $(RELEASE_DIR)/$(os)/$(arch)/$(BINARY_NAME)$(ext) .
	@echo "Done!"

# Build compiles the Go code and outputs the binary into the build directory.
build:
	@echo "Building binary..."
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
vet: tidy
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

# Verify all the module dependencies.
verify:
	@echo "Verifying the module dependencies..."
	$(GOMOD) verify

# Perform a sanity check.
check: fmt-check test vet verify build
	@echo "Check complete!"

# Format the Go code or produces an error.
fmt-check:
	@echo "Checking code formatting..."
	$(GOFMT) ./... | tee /dev/stderr | (! read)
