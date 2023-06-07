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

.PHONY: clean
clean:  ## Clean build bundles
	rm -f ./feishu2md

.PHONY: format
format:
	gofmt -l -w .
