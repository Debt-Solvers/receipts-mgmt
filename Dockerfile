# Stage 1: Build the application
FROM golang:1.23.2 AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all files and build the binary
COPY . .
RUN go build -o receipts-service ./cmd/receipt-service/main.go

# Stage 2: Minimal runtime with `glibc` support
FROM frolvlad/alpine-glibc:3.17

WORKDIR /app

# Copy application binary and configuration files
COPY --from=builder /app/receipts-service .
COPY configs/config.yaml /app/configs/config.yaml
COPY --from=builder /app/db /app/db

EXPOSE 8082
ENV PORT=8082

# Health check (optional)
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:$PORT/health || exit 1

CMD ["./receipts-service"]
