ARG VERSION=V1
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /build
RUN apk add --no-cache git ca-certificates
COPY ${VERSION}/go.mod ${VERSION}/go.sum ./
RUN go mod download
COPY ${VERSION}/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /build/fehmiobject ./cmd/server
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S fehmiobject \
    && adduser -S -G fehmiobject fehmiobject \
    && mkdir -p /app/configs /data \
    && chown -R fehmiobject:fehmiobject /app /data
WORKDIR /app
COPY --from=builder /build/fehmiobject /app/fehmiobject
COPY --chown=fehmiobject:fehmiobject configs/ /app/configs/
COPY --chown=fehmiobject:fehmiobject migrations/ /app/migrations/
USER fehmiobject
EXPOSE 9000 9001
ENTRYPOINT ["/app/fehmiobject"]
