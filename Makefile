BINARY := prwlrctl
BUILD_DIR := bin

GO := go
GOFLAGS := -trimpath -buildvcs=false

# Inject the current git tag/describe into the CLI for `--version`.
# Falls back to "dev" when building outside a git checkout.
VERSION ?= $(shell (git describe --tags --dirty --always 2>/dev/null || echo dev))

LDFLAGS := -s -w -buildid= \
	-X github.com/r3drun3/prwlrctl/cmd.version=$(VERSION)

.PHONY: build test lint fmt install clean

build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build \
		$(GOFLAGS) \
		-ldflags="$(LDFLAGS)" \
		-tags="netgo osusergo" \
		-o $(BUILD_DIR)/$(BINARY) .

install:
	CGO_ENABLED=0 $(GO) install \
		$(GOFLAGS) \
		-ldflags="$(LDFLAGS)" \
		-tags="netgo osusergo" .

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf $(BUILD_DIR)
