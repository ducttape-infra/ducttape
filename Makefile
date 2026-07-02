BUILD_DIR ?= out
BINARY_NAME ?= ducttape

GO_BUILDFLAGS ?= -buildvcs=false
LDFLAGS ?= -s -w -X main.version=$(shell cat VERSION)


.PHONY: all
all: build

.PHONY: download-gvproxy
download-gvproxy:
	mkdir -p cmd/ducttape/assets
	curl -sL "$(GVPROXY_DOWNLOAD_BASEURL)/gvproxy-linux-amd64" -o cmd/ducttape/assets/gvproxy
	chmod +x cmd/ducttape/assets/gvproxy

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor: tidy
	go mod vendor

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) cmd/ducttape/assets/

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: build
build: $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ducttape/

.PHONY: cross
cross: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME) ./cmd/ducttape/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/$(BINARY_NAME) ./cmd/ducttape/

.PHONY: test
test:
	go test ./...
