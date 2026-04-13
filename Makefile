.DEFAULT_GOAL := build
HAS_UPX := $(shell command -v upx 2> /dev/null)

.PHONY: build
build:  ## Build for current platform (cross-compile support: linux-amd64, darwin-arm64)
	go build -ldflags="-X main.version=v2-`git rev-parse --short HEAD`" -o ./feishu2md cmd/*.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./feishu2md
endif

.PHONY: build-linux-amd64
build-linux-amd64:  ## Build for Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=v2-$$(git rev-parse --short HEAD)" -o ./feishu2md-linux-amd64 cmd/*.go

.PHONY: build-darwin-arm64
build-darwin-arm64:  ## Build for macOS ARM64
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=v2-$$(git rev-parse --short HEAD)" -o ./feishu2md-darwin-arm64 cmd/*.go

.PHONY: build-all-platforms
build-all-platforms: build build-linux-amd64 build-darwin-arm64  ## Build for all platforms

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
