FROM golang:1.25-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=arm64

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY vendor ./vendor

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /out/memoir ./cmd/memoir

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/memoir /app/memoir
COPY config.json /app/config.json

EXPOSE 8080

ENTRYPOINT ["/app/memoir"]
CMD ["-config", "/app/config.json"]
