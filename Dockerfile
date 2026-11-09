# syntax=docker/dockerfile:1
# ------------------------------------------------------------------------------
# Stage 1: Build Enterprise Web UI Dashboard
# ------------------------------------------------------------------------------
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --no-audit --prefer-offline
COPY web/ ./
RUN npm run build

# ------------------------------------------------------------------------------
# Stage 2: Build Refweir Enterprise Daemon and CLI
# ------------------------------------------------------------------------------
FROM golang:1.26-alpine AS go-builder
ARG VERSION=v0.6.0-beta
WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X refweir.dev/internal/cli.Version=${VERSION}" -o /bin/refweird ./cmd/refweird && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X refweir.dev/internal/cli.Version=${VERSION}" -o /bin/refweir ./cmd/refweir

# ------------------------------------------------------------------------------
# Stage 3: Hardened, Non-Root Enterprise Production Container
# ------------------------------------------------------------------------------
FROM debian:bookworm-slim

# Install runtime dependencies: git (incl. git-http-backend) is required for Smart-HTTP transport
RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates tzdata wget \
    && rm -rf /var/lib/apt/lists/*

# Create unprivileged service user and group
RUN groupadd -g 1000 refweir && \
    useradd -u 1000 -g refweir -s /bin/sh -m refweir

# Create persistent storage directories with strict access permissions
RUN mkdir -p /var/lib/refweir/git /var/lib/refweir/audit && \
    chown -R refweir:refweir /var/lib/refweir && \
    chmod 700 /var/lib/refweir

# Copy binaries from builder
COPY --from=go-builder /bin/refweird /usr/local/bin/refweird
COPY --from=go-builder /bin/refweir /usr/local/bin/refweir

# Copy static Web UI assets into daemon directory
COPY --from=web-builder /app/internal/webui/dist /var/lib/refweir/web/dist

USER refweir:refweir
WORKDIR /var/lib/refweir

ENV REFWEIR_DATA_DIR=/var/lib/refweir \
    REFWEIR_LISTEN_ADDR=0.0.0.0:8080 \
    REFWEIR_SSH_ADDR=0.0.0.0:2222 \
    REFWEIR_PUBLIC_URL=http://localhost:8080 \
    REFWEIR_GIT_URL=localhost:8080

EXPOSE 8080 2222

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/refweird"]
