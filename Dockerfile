# Stage 1: Build Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o impersonate-service \
    .

# Stage 2: Download curl-impersonate binaries
FROM alpine:latest AS downloader

RUN apk add --no-cache wget tar ca-certificates

WORKDIR /tmp

# Download pre-compiled binaries
RUN wget -q https://github.com/lwthiker/curl-impersonate/releases/download/v0.6.1/curl-impersonate-v0.6.1.x86_64-linux-gnu.tar.gz && \
    tar -xzf curl-impersonate-v0.6.1.x86_64-linux-gnu.tar.gz

# Stage 3: Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    nss \
    nss-tools \
    libstdc++ \
    libgcc \
    wget

# Copy Go service binary
COPY --from=builder /build/impersonate-service /usr/local/bin/impersonate-service

# Copy curl-impersonate binaries
COPY --from=downloader /tmp/curl-impersonate-chrome /usr/local/bin/curl-impersonate-chrome
COPY --from=downloader /tmp/curl-impersonate-ff /usr/local/bin/curl-impersonate-ff
COPY --from=downloader /tmp/libcurl-impersonate-chrome.so* /usr/local/lib/
COPY --from=downloader /tmp/libcurl-impersonate-ff.so* /usr/local/lib/

# Make binaries executable
RUN chmod +x /usr/local/bin/curl-impersonate-chrome /usr/local/bin/curl-impersonate-ff

# Copy browsers.json
COPY browsers.json /etc/impersonate/browsers.json

# Set library path
ENV LD_LIBRARY_PATH=/usr/local/lib

# Service configuration
ENV PORT=8080
ENV BROWSERS_JSON_PATH=/etc/impersonate/browsers.json

EXPOSE 8080

# Create non-root user
RUN adduser -D -u 1000 impersonate
USER impersonate

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/impersonate-service"]
