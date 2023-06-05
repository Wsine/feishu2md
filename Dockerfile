ARG GO_VERSION=1.17

FROM golang:${GO_VERSION}-alpine AS builder

ENV GIN_MODE=release

WORKDIR /feishu2md

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o ./app ./web/*.go

FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates

ENV GIN_MODE=release

WORKDIR /feishu2md
COPY --from=builder /feishu2md/app .
COPY ./web/templ ./web/templ

EXPOSE 8080

ENTRYPOINT ["./app"]
