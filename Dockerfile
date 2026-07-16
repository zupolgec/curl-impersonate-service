# Stage 1: Download curl-impersonate binaries and libraries
FROM alpine:latest AS downloader

RUN apk add --no-cache wget tar ca-certificates

WORKDIR /tmp

# Download pre-compiled binaries and libraries for the target architecture.
# Uses the actively-maintained lexiforest fork (lwthiker's is archived).
ARG CURL_IMPERSONATE_VERSION=v1.2.2
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then ARCH="aarch64"; else ARCH="x86_64"; fi && \
    BASE="https://github.com/lexiforest/curl-impersonate/releases/download/${CURL_IMPERSONATE_VERSION}" && \
    URL_BIN="${BASE}/curl-impersonate-${CURL_IMPERSONATE_VERSION}.${ARCH}-linux-gnu.tar.gz" && \
    URL_LIB="${BASE}/libcurl-impersonate-${CURL_IMPERSONATE_VERSION}.${ARCH}-linux-gnu.tar.gz" && \
    wget -q $URL_BIN -O bin.tar.gz && tar -xzf bin.tar.gz && rm bin.tar.gz && \
    wget -q $URL_LIB -O lib.tar.gz && mkdir -p lib && tar -xzf lib.tar.gz -C lib && rm lib.tar.gz

# Stage 2: Build Go binary (Use Debian for GLIBC compatibility)
FROM golang:1.26.5-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    pkg-config \
    libcurl4-openssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy libraries from downloader for linking
COPY --from=downloader /tmp/lib/* /usr/local/lib/

# Build the binary for the target platform
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} \
    CGO_LDFLAGS="-L/usr/local/lib -lcurl-impersonate" \
    go build \
    -ldflags="-w -s" \
    -o impersonate-service \
    .

# Stage 3: Final runtime image
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnss3 \
    libnss3-tools \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Copy Go service binary
COPY --from=builder /build/impersonate-service /usr/local/bin/impersonate-service

# Copy the unified curl-impersonate binary and the wrapper scripts
COPY --from=downloader /tmp/curl-impersonate /usr/local/bin/curl-impersonate
COPY --from=downloader /tmp/curl_* /usr/local/bin/
# Copy the libs for runtime linking
COPY --from=downloader /tmp/lib/libcurl-impersonate* /usr/local/lib/
# Refresh ldconfig to find the new libs
RUN ldconfig

# Make all executables executable
RUN chmod +x /usr/local/bin/curl-impersonate /usr/local/bin/curl_*

# Copy browsers.json
COPY browsers.json /etc/impersonate/browsers.json

# Service configuration
ENV PORT=8080
ENV BROWSERS_JSON_PATH=/etc/impersonate/browsers.json
ENV DATA_DIR=/data

EXPOSE 8080

# Create non-root user and a writable data dir for the SQLite datastore
RUN useradd -m -u 1000 -s /bin/bash impersonate && \
    mkdir -p /data && chown 1000:1000 /data
VOLUME /data
USER impersonate

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/impersonate-service"]
