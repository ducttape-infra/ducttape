BUILD_DIR ?= out
BINARY_NAME ?= machine

GO_BUILDFLAGS ?= -buildvcs=false
LDFLAGS ?= -s -w

GVPROXY_VERSION ?= v0.8.7
GVPROXY_DOWNLOAD_BASEURL := https://github.com/containers/gvisor-tap-vsock/releases/download/$(GVPROXY_VERSION)

.PHONY: all
all: build

.PHONY: download-gvproxy
download-gvproxy:
	mkdir -p cmd/machine/assets
	curl -sL "$(GVPROXY_DOWNLOAD_BASEURL)/gvproxy-linux-amd64" -o cmd/machine/assets/gvproxy
	chmod +x cmd/machine/assets/gvproxy

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor: tidy
	go mod vendor

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) cmd/machine/assets/

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: build
build: download-gvproxy tidy $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/machine/

.PHONY: cross
cross: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME) ./cmd/machine/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/$(BINARY_NAME) ./cmd/machine/

.PHONY: test
test:
	go test ./...
