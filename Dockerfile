# Stage 1: Download curl-impersonate binaries
FROM alpine:latest AS downloader

RUN apk add --no-cache wget tar ca-certificates

WORKDIR /tmp

# Download pre-compiled binaries for the target architecture
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        wget -q https://github.com/lwthiker/curl-impersonate/releases/download/v0.6.1/curl-impersonate-v0.6.1.aarch64-linux-gnu.tar.gz && \
        tar -xzf curl-impersonate-v0.6.1.aarch64-linux-gnu.tar.gz; \
    else \
        wget -q https://github.com/lwthiker/curl-impersonate/releases/download/v0.6.1/curl-impersonate-v0.6.1.x86_64-linux-gnu.tar.gz && \
        tar -xzf curl-impersonate-v0.6.1.x86_64-linux-gnu.tar.gz; \
    fi

# Stage 2: Build Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev pkgconfig

# Stage 3: Final runtime image
# Use Debian slim for glibc compatibility (curl-impersonate binaries need glibc)
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

# Copy curl-impersonate binaries and wrapper scripts
COPY --from=downloader /tmp/curl-impersonate-chrome /usr/local/bin/curl-impersonate-chrome
COPY --from=downloader /tmp/curl-impersonate-ff /usr/local/bin/curl-impersonate-ff
COPY --from=downloader /tmp/curl_* /usr/local/bin/

# Make all executables executable
RUN chmod +x /usr/local/bin/curl-impersonate-chrome /usr/local/bin/curl-impersonate-ff /usr/local/bin/curl_*

# Copy browsers.json
COPY browsers.json /etc/impersonate/browsers.json

# Service configuration
ENV PORT=8080
ENV BROWSERS_JSON_PATH=/etc/impersonate/browsers.json

EXPOSE 8080

# Create non-root user (Debian syntax)
RUN useradd -m -u 1000 -s /bin/bash impersonate
USER impersonate

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/impersonate-service"]
