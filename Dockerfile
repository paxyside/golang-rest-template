FROM golang:1.23.6 AS build

LABEL builder="golangci-lint & go build"

WORKDIR /build

RUN curl -sSfL https://github.com/golangci/golangci-lint/releases/download/v2.1.5/golangci-lint-2.1.5-linux-amd64.tar.gz | tar -xz -C /usr/local/bin

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY docker .
RUN golangci-lint run --fix --timeout=10m
RUN go test -v ./...
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o /usr/bin/server ./cmd/server/main.go

FROM alpine:3.12

RUN apk update && \
    apk add --no-cache \
        ca-certificates \
        curl \
        tzdata \
    && rm -rf -- /var/cache/apk/*

ENV TZ="UTC"
WORKDIR /app
COPY --from=build /usr/bin/server .

HEALTHCHECK --interval=20s --timeout=5s --retries=5 --start-period=30s \
    CMD curl -fsS -m5 -A'docker-healthcheck' http://127.0.0.1/api/ping || exit 1

CMD ["./server"]
