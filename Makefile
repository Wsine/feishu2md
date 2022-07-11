.DEFAULT_GOAL := build-linux
HAS_UPX := $(shell command -v upx 2> /dev/null)

.PHONY: all
all: build-darwin build-linux build-windows ## Build all platforms

.PHONY: build-darwin
build-darwin: ## Build for MacOS
	rm -rf ./bin/darwin
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags "-s -w" -o ./bin/darwin/feishu2md *.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./bin/darwin/feishu2md
endif

.PHONY: build-linux
build-linux: ## Build for Linux
	rm -rf ./bin/linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags "-s -w" -o ./bin/linux/feishu2md *.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./bin/linux/feishu2md
endif

.PHONY: build-windows
build-windows: ## Build for Windows
	rm -rf ./bin/windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags "-s -w" -o ./bin/windows/feishu2md.exe *.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./bin/windows/feishu2md.exe
endif

.PHONY: clean
clean:  ## Clean build bundles
	rm -rf ./bin

.PHONY: pack
pack:  # Pack up build bundles
	zip -r feishu2md-amd64.zip ./bin
