# Variables from the original YAML
BINARY_NAME        ?= quantifiable-swap
OKX_API_KEY        ?=
OKX_SECRET_KEY     ?=
OKX_PASSPHRASE     ?=
BARK_TOKEN         ?=

REMOTE_PM2_PATH    ?=
REMOTE_ADDRESS     ?=
REMOTE_PATH        ?=
PROXY_ADDR         ?=

# Default targets
.PHONY: default dev clean build linux-amd64 darwin-amd64

default: clean linux-amd64
dev: clean darwin-amd64

# Clean and prepare build directory
clean:
	rm -rf dist
	mkdir -p dist

# Build task with Go build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) GOMIPS=$(GOMIPS) CGO_ENABLED=0 \
	go build -trimpath -o ./dist/$(BINARY_NAME) \
	  -ldflags "\
		-X 'github.com/gtoxlili/quantifiable-swap/constants.OkxAPIKey=$(OKX_API_KEY)' \
		-X 'github.com/gtoxlili/quantifiable-swap/constants.OkxSecretKey=$(OKX_SECRET_KEY)' \
		-X 'github.com/gtoxlili/quantifiable-swap/constants.OkxPassphrase=$(OKX_PASSPHRASE)' \
		-X 'github.com/gtoxlili/quantifiable-swap/constants.ProxyAddr=$(PROXY_ADDR)' \
		-X 'github.com/gtoxlili/quantifiable-swap/constants.BarkToken=$(BARK_TOKEN)' \
		-w -s -buildid="

# Linux build and deployment
linux-amd64:
	$(MAKE) build GOOS=linux GOARCH=amd64 PROXY_ADDR=$(PROXY_ADDR)
	ssh $(REMOTE_ADDRESS) "$(REMOTE_PM2_PATH) delete $(BINARY_NAME) || true"
	scp -r ./dist/$(BINARY_NAME) $(REMOTE_ADDRESS):$(REMOTE_PATH)
	ssh $(REMOTE_ADDRESS) "$(REMOTE_PM2_PATH) start $(REMOTE_PATH)$(BINARY_NAME)"

# macOS build
darwin-amd64:
	$(MAKE) build GOOS=darwin GOARCH=amd64
