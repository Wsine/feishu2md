.DEFAULT_GOAL := build
HAS_UPX := $(shell command -v upx 2> /dev/null)

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags "-s -w" -o ./bin/feishu2md main.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./bin/feishu2md
endif

.PHONY: clean
clean:  ## Clean build bundles
	rm -rf ./bin
