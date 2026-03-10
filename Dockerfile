# Dockerfile for Member API
# Build with: docker build -t member-api .

FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o member_api ./main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates wget && \
    apk --no-cache upgrade libssl3 libcrypto3 && \
    update-ca-certificates
WORKDIR /app

COPY --from=builder /src/member_api ./member_api

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./member_api"]
