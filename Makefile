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
