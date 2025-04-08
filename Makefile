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

# Set default target OS's if none are specified.
TARGETS ?= darwin linux windows

# Set the build architecture.
ARCH ?= amd64 arm64

.PHONY: all release build test run fmt vet clean tidy verify install release-all checkout-master

checkout-master:
	@echo "Checking out the master branch"
	git checkout master

install: checkout-master
	@echo "Installing $(BINARY_NAME) for your system"
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOINSTALL) -trimpath -ldflags="-s -w" .
	@echo "Done! Don't forget to create your configuration file (see README.md) before running."

uninstall:
	rm "$(GOPATH)/bin/$(BINARY_NAME)"
	@echo "Done! Don't forget to delete any configuration file you created."

create-build-dir:
	@mkdir -p $(BUILD_DIR)

release-all: checkout-master vet verify test clean create-build-dir
	@echo "Building binaries for: $(TARGETS)"
	@for os in $(TARGETS); do \
		for arch in $(ARCH); do \
			echo "Building for $$os $$arch..."; \
			ext=""; \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			fi; \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$$os/$$arch/$(BINARY_NAME)$$ext .; \
		done; \
	done

release: checkout-master vet verify test clean create-build-dir
	@os="$(GOOS)"
	@arch="$(GOARCH)"
	@ext="";
	@if [ "$$os" = "windows" ]; then \
		ext=".exe"; \
	fi;
	GOOS=$$os GOARCH=$$arch $(GOBUILD) -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$$os/$$arch/$(BINARY_NAME)$$ext .;

build: create-build-dir
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

# Verify all the module dependencies
verify:
	@echo "Verifying the module dependencies..."
	$(GOMOD) verify
