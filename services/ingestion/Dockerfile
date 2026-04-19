FROM golang:1.26-alpine AS builder

# libheif-dev + build tools needed for CGO
RUN apk add --no-cache libheif-dev gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o ingestion-service .

FROM alpine:3.19

# Runtime libheif dependency
RUN apk add --no-cache libheif ca-certificates

WORKDIR /app
COPY --from=builder /app/ingestion-service .

EXPOSE 8080

CMD ["./ingestion-service"]
