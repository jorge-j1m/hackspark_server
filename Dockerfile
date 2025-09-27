# Build stage
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -a -installsuffix cgo \
    -o hackspark-api ./cmd/api

# Final stage - use distroless for better security
FROM gcr.io/distroless/static:nonroot

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder stage
COPY --from=builder /app/hackspark-api /usr/local/bin/hackspark-api

# Use non-root user (distroless nonroot user has UID 65532)
USER 65532:65532

# Expose port
EXPOSE 8080

# Set environment variables
ENV ENVIRONMENT=production

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/hackspark-api", "--health-check"] || exit 1

# Command to run
ENTRYPOINT ["/usr/local/bin/hackspark-api"]