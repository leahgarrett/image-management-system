# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend .

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Copy binary from builder
COPY --from=builder /app/backend .

EXPOSE 8081

ENTRYPOINT ["./backend"]
