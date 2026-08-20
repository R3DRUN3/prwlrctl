BINARY := prwlrctl
BUILD_DIR := bin

GO := go
GOFLAGS := -trimpath -buildvcs=false
LDFLAGS := -s -w -buildid=

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