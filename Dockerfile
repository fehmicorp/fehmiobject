ARG VERSION=v1

# ==========================================
# Build stage
# ==========================================
FROM golang:1.26-alpine AS builder
ARG VERSION
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
RUN apk add --no-cache git ca-certificates
COPY ${VERSION}/go.mod ${VERSION}/go.sum ./
RUN go mod download
COPY ${VERSION}/ ./
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /build/blobstorage .

# ==========================================
# Runtime stage
# ==========================================
FROM alpine:3.22
ARG VERSION
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S blobstorage \
    && adduser -S -G blobstorage blobstorage \
    && mkdir -p /app /app/configs /data/hot /data/persistent /data/staging \
    && chown -R blobstorage:blobstorage /app /data
WORKDIR /app
COPY --from=builder --chown=blobstorage:blobstorage /build/blobstorage /app/blobstorage
COPY --chown=blobstorage:blobstorage configs/${VERSION}.yaml /app/config.yaml
USER blobstorage
EXPOSE 9000 9001
VOLUME ["/data/hot", "/data/persistent", "/data/staging"]
ENTRYPOINT ["/app/blobstorage"]
CMD ["-config", "/app/config.yaml"]
