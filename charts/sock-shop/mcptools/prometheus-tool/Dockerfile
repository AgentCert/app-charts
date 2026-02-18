# Multi-stage build for optimal image size
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go modules)
RUN apk add --no-cache git

# Copy go module files
COPY go.mod ./

# Copy source code
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o prometheus-tool .

# Final stage - minimal runtime image
FROM alpine:3.18

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/prometheus-tool .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8083

# Set environment variables
ENV PORT=8083
ENV PROMETHEUS_URL=http://prometheus.monitoring.svc.cluster.local:9090

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:8083/health || exit 1

# Run the binary
ENTRYPOINT ["./prometheus-tool"]
