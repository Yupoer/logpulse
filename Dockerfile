# ==========================================
# Stage 1: Builder
# ==========================================
FROM golang:1.25-alpine AS builder

# set working directory
WORKDIR /app

# 1. Download dependencies (use Docker Layer Caching)
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source code and compile
COPY . .
# CGO_ENABLED=0: close CGO, ensure static binary
# GOOS=linux: force compile for Linux (because container is running on Linux)
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o logpulse cmd/api/main.go

# ==========================================
# Stage 2: Runner
# Linux (Alpine)
# ==========================================
FROM alpine:latest

WORKDIR /root/

# install necessary base packages (HTTPS certificates & timezone data)
RUN apk --no-cache add ca-certificates tzdata

# set timezone (Taipei)
ENV TZ=Asia/Taipei

# copy the compiled "logpulse" executable from the Builder layer
# note: only copy the executable
COPY --from=builder /app/logpulse .

# declare port (for documentation purposes, actually mapped in docker-compose)
EXPOSE 8080

# start the application
CMD ["./logpulse"]