.DEFAULT_GOAL := build
HAS_UPX := $(shell command -v upx 2> /dev/null)

.PHONY: build
build:
	go build -ldflags="-X main.version=v2-`git rev-parse --short HEAD`" -o ./feishu2md cmd/*.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./feishu2md
endif

.PHONY: test
test:
	go test ./...

.PHONY: server
server:
	go build -o ./feishu2md4web web/*.go

.PHONY: image
image:
	docker build -t feishu2md .

.PHONY: docker
docker:
	docker run -it --rm -p 8080:8080 feishu2md

.PHONY: clean
clean:  ## Clean build bundles
	rm -f ./feishu2md ./feishu2md4web

.PHONY: format
format:
	gofmt -l -w .

.PHONY: all
all: build server
	@echo "Build all done"

# 跨平台编译配置
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v2-dev")
LDFLAGS := -s -w -X main.version=$(VERSION)
OUTPUT_DIR := dist

.PHONY: cross-build
cross-build: clean-dist
	@echo "Building for all platforms..."
	@mkdir -p $(OUTPUT_DIR)
	# Linux amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-linux-amd64 cmd/*.go
	# Linux arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-linux-arm64 cmd/*.go
	# Windows amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-windows-amd64.exe cmd/*.go
	# Darwin amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-darwin-amd64 cmd/*.go
	# Darwin arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-darwin-arm64 cmd/*.go
	@echo "Cross-build completed! Binaries in $(OUTPUT_DIR)/"

.PHONY: cross-build-linux
cross-build-linux: clean-dist
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-linux-amd64 cmd/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-linux-arm64 cmd/*.go

.PHONY: cross-build-windows
cross-build-windows: clean-dist
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-windows-amd64.exe cmd/*.go

.PHONY: cross-build-darwin
cross-build-darwin: clean-dist
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-darwin-amd64 cmd/*.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/feishu2md-darwin-arm64 cmd/*.go

.PHONY: clean-dist
clean-dist:
	rm -rf $(OUTPUT_DIR)
