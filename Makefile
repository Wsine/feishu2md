.DEFAULT_GOAL := build
HAS_UPX := $(shell command -v upx 2> /dev/null)

.PHONY: build
build:
	go build -o ./bin/feishu2md cmd/main.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./bin/feishu2md
endif

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:  ## Clean build bundles
	rm -rf ./bin

.PHONY: format
format:
	gofmt -l -w .
