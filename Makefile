BINARY  := googledrive-mcp
CMD     := ./cmd/googledrive-mcp
DIST    := dist
VERSION ?= dev

# -s -w strips debug info; -X injects the version (see version-embedded-via-ldflags).
LDFLAGS := -s -w -X main.version=$(VERSION)

# Canonical cross-compile targets. Kept in sync with internal/build.SupportedTargets
# by TestMakefilePlatformsMatchSupportedTargets — edit both together.
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: build build-all clean test

# Build a single binary for the host platform. The project is pure Go, so
# CGO_ENABLED=0 produces a static binary with no dynamic-library dependencies.
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

# Cross-compile every supported platform into $(DIST). -trimpath keeps the
# output reproducible so the build is idempotent from CI or a local checkout.
build-all: clean
	@mkdir -p $(DIST)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		out=$(DIST)/$(BINARY)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then out=$$out.exe; fi; \
		echo "building $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -trimpath -ldflags "$(LDFLAGS)" -o $$out $(CMD) || exit 1; \
	done

clean:
	@rm -rf $(DIST)

test:
	go test ./...
