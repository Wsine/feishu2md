.DEFAULT_GOAL := build
HAS_UPX := $(shell command -v upx 2> /dev/null)

.PHONY: build
build:
		go build -o ./bin/feishu2md main.go
ifneq ($(and $(COMPRESS),$(HAS_UPX)),)
	upx -9 ./bin/feishu2md
endif

.PHONY: clean
clean:  ## Clean build bundles
	rm -rf ./bin
